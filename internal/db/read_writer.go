package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/watchtower/internal/db/common"
	"github.com/hashicorp/watchtower/internal/oplog"
	"github.com/hashicorp/watchtower/internal/oplog/store"
	"github.com/jinzhu/gorm"
	"google.golang.org/protobuf/proto"
)

const (
	NoRowsAffected = 0

	// DefaultLimit is the default for results for watchtower
	DefaultLimit = 10000
)

// Reader interface defines lookups/searching for resources
type Reader interface {
	// LookupByName will lookup resource by its friendly name which must be unique
	LookupByName(ctx context.Context, resource ResourceNamer, opt ...Option) error

	// LookupByPublicId will lookup resource by its public_id which must be unique
	LookupByPublicId(ctx context.Context, resource ResourcePublicIder, opt ...Option) error

	// LookupWhere will lookup and return the first resource using a where clause with parameters
	LookupWhere(ctx context.Context, resource interface{}, where string, args ...interface{}) error

	// SearchWhere will search for all the resources it can find using a where
	// clause with parameters. Supports the WithLimit option.  If
	// WithLimit < 0, then unlimited results are returned.  If WithLimit == 0, then
	// default limits are used for results.
	SearchWhere(ctx context.Context, resources interface{}, where string, args []interface{}, opt ...Option) error

	// ScanRows will scan sql rows into the interface provided
	ScanRows(rows *sql.Rows, result interface{}) error

	// DB returns the sql.DB
	DB() (*sql.DB, error)
}

// Writer interface defines create, update and retryable transaction handlers
type Writer interface {
	// DoTx will wrap the TxHandler in a retryable transaction
	DoTx(ctx context.Context, retries uint, backOff Backoff, Handler TxHandler) (RetryInfo, error)

	// Update an object in the db, fieldMask is required and provides
	// field_mask.proto paths for fields that should be updated. The i interface
	// parameter is the type the caller wants to update in the db and its
	// fields are set to the update values. setToNullPaths is optional and
	// provides field_mask.proto paths for the fields that should be set to
	// null.  fieldMaskPaths and setToNullPaths must not intersect. The caller
	// is responsible for the transaction life cycle of the writer and if an
	// error is returned the caller must decide what to do with the transaction,
	// which almost always should be to rollback.  Update returns the number of
	// rows updated or an error. Supported options: WithOplog.
	Update(ctx context.Context, i interface{}, fieldMaskPaths []string, setToNullPaths []string, opt ...Option) (int, error)

	// Create an object in the db with options: WithOplog
	// the caller is responsible for the transaction life cycle of the writer
	// and if an error is returned the caller must decide what to do with
	// the transaction, which almost always should be to rollback.
	Create(ctx context.Context, i interface{}, opt ...Option) error

	// Delete an object in the db with options: WithOplog
	// the caller is responsible for the transaction life cycle of the writer
	// and if an error is returned the caller must decide what to do with
	// the transaction, which almost always should be to rollback. Delete
	// returns the number of rows deleted or an error.
	Delete(ctx context.Context, i interface{}, opt ...Option) (int, error)

	// DB returns the sql.DB
	DB() (*sql.DB, error)
}

const (
	StdRetryCnt = 20
)

// RetryInfo provides information on the retries of a transaction
type RetryInfo struct {
	Retries int
	Backoff time.Duration
}

// TxHandler defines a handler for a func that writes a transaction for use with DoTx
type TxHandler func(Reader, Writer) error

// ResourcePublicIder defines an interface that LookupByPublicId() can use to get the resource's public id
type ResourcePublicIder interface {
	GetPublicId() string
}

// ResourceNamer defines an interface that LookupByName() can use to get the resource's friendly name
type ResourceNamer interface {
	GetName() string
}

type OpType int

const (
	UnknownOp OpType = 0
	CreateOp  OpType = 1
	UpdateOp  OpType = 2
	DeleteOp  OpType = 3
)

// VetForWriter provides an interface that Create and Update can use to vet the
// resource before before writing it to the db.  For optType == UpdateOp,
// options WithFieldMaskPath and WithNullPaths are supported.  For optType ==
// CreateOp, no options are supported
type VetForWriter interface {
	VetForWrite(ctx context.Context, r Reader, opType OpType, opt ...Option) error
}

// Db uses a gorm DB connection for read/write
type Db struct {
	underlying *gorm.DB
}

// ensure that Db implements the interfaces of: Reader and Writer
var _ Reader = (*Db)(nil)
var _ Writer = (*Db)(nil)

func New(underlying *gorm.DB) *Db {
	return &Db{underlying: underlying}
}

// DB returns the sql.DB
func (rw *Db) DB() (*sql.DB, error) {
	if rw.underlying == nil {
		return nil, fmt.Errorf("missing underlying db: %w", ErrNilParameter)
	}
	return rw.underlying.DB(), nil
}

// Scan rows will scan the rows into the interface
func (rw *Db) ScanRows(rows *sql.Rows, result interface{}) error {
	if rw.underlying == nil {
		return fmt.Errorf("scan rows: missing underlying db %w", ErrNilParameter)
	}
	if isNil(result) {
		return fmt.Errorf("scan rows: result is missing %w", ErrNilParameter)
	}
	return rw.underlying.ScanRows(rows, result)
}

func (rw *Db) lookupAfterWrite(ctx context.Context, i interface{}, opt ...Option) error {
	opts := GetOpts(opt...)
	withLookup := opts.withLookup

	if !withLookup {
		return nil
	}
	if _, ok := i.(ResourcePublicIder); ok {
		if err := rw.LookupByPublicId(ctx, i.(ResourcePublicIder), opt...); err != nil {
			return fmt.Errorf("lookup after write: failed %w", err)
		}
		return nil
	}
	return errors.New("not a resource with an id")
}

// Create an object in the db with options: WithOplog and WithLookup (to force a
// lookup after create).
func (rw *Db) Create(ctx context.Context, i interface{}, opt ...Option) error {
	if rw.underlying == nil {
		return fmt.Errorf("create: missing underlying db %w", ErrNilParameter)
	}
	if isNil(i) {
		return fmt.Errorf("create: interface is missing %w", ErrNilParameter)
	}
	opts := GetOpts(opt...)
	withOplog := opts.withOplog
	if withOplog {
		// let's validate oplog options before we start writing to the database
		_, err := validateOplogArgs(i, opts)
		if err != nil {
			return fmt.Errorf("create: oplog validation failed %w", err)
		}
	}
	// these fields should be nil, since they are not writeable and we want the
	// db to manage them
	setFieldsToNil(i, []string{"CreateTime", "UpdateTime"})

	if vetter, ok := i.(VetForWriter); ok {
		if err := vetter.VetForWrite(ctx, rw, CreateOp); err != nil {
			return fmt.Errorf("create: vet for write failed %w", err)
		}
	}
	var ticket *store.Ticket
	if withOplog {
		var err error
		ticket, err = rw.getTicket(i)
		if err != nil {
			return fmt.Errorf("create: unable to get ticket: %w", err)
		}
	}
	if err := rw.underlying.Create(i).Error; err != nil {
		return fmt.Errorf("create: failed %w", err)
	}
	if withOplog {
		if err := rw.addOplog(ctx, CreateOp, opts, ticket, i); err != nil {
			return err
		}
	}
	if err := rw.lookupAfterWrite(ctx, i, opt...); err != nil {
		return fmt.Errorf("create: lookup error %w", err)
	}
	return nil
}

// CreateItems will create multiple items of the same type.
// Supported options: WithOplog.  WithLookup is not a supported option.
func (rw *Db) CreateItems(ctx context.Context, createItems []interface{}, opt ...Option) error {
	if rw.underlying == nil {
		return fmt.Errorf("create items: missing underlying db: %w", ErrNilParameter)
	}
	if len(createItems) == 0 {
		return fmt.Errorf("create items: no interfaces to create: %w", ErrInvalidParameter)
	}
	opts := GetOpts(opt...)
	if opts.withLookup {
		return fmt.Errorf("create items: withLookup not a supported option: %w", ErrInvalidParameter)
	}
	// verify that createItems are all the same type.
	var foundType reflect.Type
	for i, v := range createItems {
		if i == 0 {
			foundType = reflect.TypeOf(v)
		}
		currentType := reflect.TypeOf(v)
		if foundType != currentType {
			return fmt.Errorf("create items: create items contains disparate types. item %d is not a %s: %w", i, foundType.Name(), ErrInvalidParameter)
		}
	}
	var ticket *store.Ticket
	if opts.withOplog {
		_, err := validateOplogArgs(createItems[0], opts)
		if err != nil {
			return fmt.Errorf("create items: oplog validation failed: %w", err)
		}
		ticket, err = rw.getTicket(createItems[0])
		if err != nil {
			return fmt.Errorf("create items: unable to get ticket: %w", err)
		}
	}
	for _, item := range createItems {
		rw.Create(ctx, item)
	}
	if opts.withOplog {
		if err := rw.addOplogForItems(ctx, CreateOp, opts, ticket, createItems); err != nil {
			return fmt.Errorf("create items: unable to add oplog: %w", err)
		}
	}
	return nil
}

// Update an object in the db, fieldMask is required and provides
// field_mask.proto paths for fields that should be updated. The i interface
// parameter is the type the caller wants to update in the db and its
// fields are set to the update values. setToNullPaths is optional and
// provides field_mask.proto paths for the fields that should be set to
// null.  fieldMaskPaths and setToNullPaths must not intersect. The caller
// is responsible for the transaction life cycle of the writer and if an
// error is returned the caller must decide what to do with the transaction,
// which almost always should be to rollback.  Update returns the number of
// rows updated. Supported options: WithOplog and WithVersion.  If WithVersion
// is used, then the update will include the version number in the update where
// clause, which basically makes the update use optimistic locking and the
// update will only succeed if the existing rows version matches the WithVersion
// option.
func (rw *Db) Update(ctx context.Context, i interface{}, fieldMaskPaths []string, setToNullPaths []string, opt ...Option) (int, error) {
	if rw.underlying == nil {
		return NoRowsAffected, fmt.Errorf("update: missing underlying db %w", ErrNilParameter)
	}
	if isNil(i) {
		return NoRowsAffected, fmt.Errorf("update: interface is missing %w", ErrNilParameter)
	}
	if len(fieldMaskPaths) == 0 && len(setToNullPaths) == 0 {
		return NoRowsAffected, errors.New("update: both fieldMaskPaths and setToNullPaths are missing")
	}

	// we need to filter out some non-updatable fields (like: CreateTime, etc)
	fieldMaskPaths = filterPaths(fieldMaskPaths)
	setToNullPaths = filterPaths(setToNullPaths)
	if len(fieldMaskPaths) == 0 && len(setToNullPaths) == 0 {
		return NoRowsAffected, fmt.Errorf("update: after filtering non-updated fields, there are no fields left in fieldMaskPaths or setToNullPaths")
	}

	updateFields, err := common.UpdateFields(i, fieldMaskPaths, setToNullPaths)
	if err != nil {
		return NoRowsAffected, fmt.Errorf("update: getting update fields failed: %w", err)
	}
	if len(updateFields) == 0 {
		return NoRowsAffected, fmt.Errorf("update: no fields matched using fieldMaskPaths %s", fieldMaskPaths)
	}

	// This is not a watchtower scope, but rather a gorm Scope:
	// https://godoc.org/github.com/jinzhu/gorm#DB.NewScope
	scope := rw.underlying.NewScope(i)
	if scope.PrimaryKeyZero() {
		return NoRowsAffected, fmt.Errorf("update: primary key is not set")
	}

	for _, f := range scope.PrimaryFields() {
		if contains(fieldMaskPaths, f.Name) {
			return NoRowsAffected, fmt.Errorf("update: not allowed on primary key field %s: %w", f.Name, ErrInvalidFieldMask)
		}
	}

	opts := GetOpts(opt...)
	withOplog := opts.withOplog
	if withOplog {
		// let's validate oplog options before we start writing to the database
		_, err := validateOplogArgs(i, opts)
		if err != nil {
			return NoRowsAffected, fmt.Errorf("update: oplog validation failed %w", err)
		}
	}
	if vetter, ok := i.(VetForWriter); ok {
		if err := vetter.VetForWrite(ctx, rw, UpdateOp, WithFieldMaskPaths(fieldMaskPaths), WithNullPaths(setToNullPaths)); err != nil {
			return NoRowsAffected, fmt.Errorf("update: vet for write failed %w", err)
		}
	}
	var ticket *store.Ticket
	if withOplog {
		var err error
		ticket, err = rw.getTicket(i)
		if err != nil {
			return NoRowsAffected, fmt.Errorf("update: unable to get ticket: %w", err)
		}
	}
	var underlying *gorm.DB
	switch {
	case opts.WithVersion > 0:
		if _, ok := scope.FieldByName("version"); !ok {
			return NoRowsAffected, fmt.Errorf("update: %s does not have a version field", scope.TableName())
		}
		underlying = rw.underlying.Model(i).Where("version = ?", opts.WithVersion).Updates(updateFields)
	default:
		underlying = rw.underlying.Model(i).Updates(updateFields)
	}
	if underlying.Error != nil {
		if err == gorm.ErrRecordNotFound {
			return NoRowsAffected, fmt.Errorf("update: failed %w", ErrRecordNotFound)
		}
		return NoRowsAffected, fmt.Errorf("update: failed %w", underlying.Error)
	}
	rowsUpdated := int(underlying.RowsAffected)
	if withOplog && rowsUpdated > 0 {
		// we don't want to change the inbound slices in opts, so we'll make our
		// own copy to pass to addOplog()
		oplogFieldMasks := make([]string, len(fieldMaskPaths))
		copy(oplogFieldMasks, fieldMaskPaths)
		oplogNullPaths := make([]string, len(setToNullPaths))
		copy(oplogNullPaths, setToNullPaths)
		oplogOpts := Options{
			oplogOpts:          opts.oplogOpts,
			withOplog:          opts.withOplog,
			WithFieldMaskPaths: oplogFieldMasks,
			WithNullPaths:      oplogNullPaths,
		}
		if err := rw.addOplog(ctx, UpdateOp, oplogOpts, ticket, i); err != nil {
			return rowsUpdated, fmt.Errorf("update: add oplog failed %w", err)
		}
	}
	// we need to force a lookupAfterWrite so the resource returned is correctly initialized
	// from the db
	opt = append(opt, WithLookup(true))
	if err := rw.lookupAfterWrite(ctx, i, opt...); err != nil {
		return NoRowsAffected, fmt.Errorf("update: lookup error %w", err)
	}
	return rowsUpdated, nil
}

// Delete an object in the db with options: WithOplog (which requires
// WithMetadata, WithWrapper). Delete returns the number of rows deleted and
// any errors.
func (rw *Db) Delete(ctx context.Context, i interface{}, opt ...Option) (int, error) {
	if rw.underlying == nil {
		return NoRowsAffected, fmt.Errorf("delete: missing underlying db %w", ErrNilParameter)
	}
	if isNil(i) {
		return NoRowsAffected, fmt.Errorf("delete: interface is missing %w", ErrNilParameter)
	}
	// This is not a watchtower scope, but rather a gorm Scope:
	// https://godoc.org/github.com/jinzhu/gorm#DB.NewScope
	scope := rw.underlying.NewScope(i)
	if scope.PrimaryKeyZero() {
		return NoRowsAffected, fmt.Errorf("delete: primary key is not set")
	}
	opts := GetOpts(opt...)
	withOplog := opts.withOplog
	if withOplog {
		_, err := validateOplogArgs(i, opts)
		if err != nil {
			return NoRowsAffected, fmt.Errorf("delete: oplog validation failed %w", err)
		}
	}
	var ticket *store.Ticket
	if withOplog {
		var err error
		ticket, err = rw.getTicket(i)
		if err != nil {
			return NoRowsAffected, fmt.Errorf("delete: unable to get ticket: %w", err)
		}
	}
	underlying := rw.underlying.Delete(i)
	if underlying.Error != nil {
		return NoRowsAffected, fmt.Errorf("delete: failed %w", underlying.Error)
	}
	rowsDeleted := int(underlying.RowsAffected)
	if withOplog && rowsDeleted > 0 {
		if err := rw.addOplog(ctx, DeleteOp, opts, ticket, i); err != nil {
			return rowsDeleted, fmt.Errorf("delete: add oplog failed %w", err)
		}
	}
	return rowsDeleted, nil
}

// DeleteItems will delete multiple items of the same type.
// Supported options: WithOplog.
func (rw *Db) DeleteItems(ctx context.Context, deleteItems []interface{}, opt ...Option) (int, error) {
	if rw.underlying == nil {
		return NoRowsAffected, fmt.Errorf("delete items: missing underlying db: %w", ErrNilParameter)
	}
	if len(deleteItems) == 0 {
		return NoRowsAffected, fmt.Errorf("delete items: no interfaces to delete: %w", ErrInvalidParameter)
	}
	// verify that createItems are all the same type.
	var foundType reflect.Type
	for i, v := range deleteItems {
		if i == 0 {
			foundType = reflect.TypeOf(v)
		}
		currentType := reflect.TypeOf(v)
		if foundType != currentType {
			return NoRowsAffected, fmt.Errorf("delete items: items contain disparate types.  item %d is not a %s: %w", i, foundType.Name(), ErrInvalidParameter)
		}
	}
	opts := GetOpts(opt...)
	var ticket *store.Ticket
	if opts.withOplog {
		_, err := validateOplogArgs(deleteItems[0], opts)
		if err != nil {
			return NoRowsAffected, fmt.Errorf("delete items: oplog validation failed: %w", err)
		}
		ticket, err = rw.getTicket(deleteItems[0])
		if err != nil {
			return NoRowsAffected, fmt.Errorf("delete items: unable to get ticket: %w", err)
		}
	}
	rowsDeleted := 0
	for _, item := range deleteItems {
		// calling delete directly on the underlying db, since the writer.Delete
		// doesn't provide capabilities needed here (which is different from the
		// relationship between Create and CreateItems).
		underlying := rw.underlying.Delete(item)
		if underlying.Error != nil {
			return rowsDeleted, fmt.Errorf("delete: failed: %w", underlying.Error)
		}
		rowsDeleted += int(underlying.RowsAffected)
	}
	if opts.withOplog && rowsDeleted > 0 {
		if err := rw.addOplogForItems(ctx, DeleteOp, opts, ticket, deleteItems); err != nil {
			return rowsDeleted, fmt.Errorf("delete items: unable to add oplog: %w", err)
		}
	}
	return rowsDeleted, nil
}

func validateOplogArgs(i interface{}, opts Options) (oplog.ReplayableMessage, error) {
	oplogArgs := opts.oplogOpts
	if oplogArgs.wrapper == nil {
		return nil, fmt.Errorf("error no wrapper WithOplog: %w", ErrNilParameter)
	}
	if len(oplogArgs.metadata) == 0 {
		return nil, fmt.Errorf("error no metadata for WithOplog: %w", ErrInvalidParameter)
	}
	replayable, ok := i.(oplog.ReplayableMessage)
	if !ok {
		return nil, errors.New("error not a replayable message for WithOplog")
	}
	return replayable, nil
}

func (rw *Db) getTicketFor(aggregateName string) (*store.Ticket, error) {
	if rw.underlying == nil {
		return nil, fmt.Errorf("get ticket for %s: underlying db missing: %w", aggregateName, ErrNilParameter)
	}
	ticketer, err := oplog.NewGormTicketer(rw.underlying, oplog.WithAggregateNames(true))
	if err != nil {
		return nil, fmt.Errorf("get ticket for %s: unable to get Ticketer %w", aggregateName, err)
	}
	ticket, err := ticketer.GetTicket(aggregateName)
	if err != nil {
		return nil, fmt.Errorf("get ticket for %s: unable to get ticket %w", aggregateName, err)
	}
	return ticket, nil
}
func (rw *Db) getTicket(i interface{}) (*store.Ticket, error) {
	if rw.underlying == nil {
		return nil, fmt.Errorf("get ticket: underlying db missing: %w", ErrNilParameter)
	}
	replayable, ok := i.(oplog.ReplayableMessage)
	if !ok {
		return nil, fmt.Errorf("get ticket: not a replayable message %w", ErrInvalidParameter)
	}
	return rw.getTicketFor(replayable.TableName())
}

// addOplogForItems will add a multi-message oplog entry with one msg for each
// item. Items must all be of the same type.  Only CreateOp and DeleteOp are
// currently supported operations.
func (rw *Db) addOplogForItems(ctx context.Context, opType OpType, opts Options, ticket *store.Ticket, items []interface{}) error {
	oplogArgs := opts.oplogOpts
	if ticket == nil {
		return fmt.Errorf("oplog many: ticket is missing: %w", ErrNilParameter)
	}
	if items == nil {
		return fmt.Errorf("oplog many: items are missing: %w", ErrNilParameter)
	}
	if len(items) == 0 {
		return fmt.Errorf("oplog many: items is empty: %w", ErrInvalidParameter)
	}
	if oplogArgs.metadata == nil {
		return fmt.Errorf("oplog many: metadata is missing: %w", ErrNilParameter)
	}
	if oplogArgs.wrapper == nil {
		return fmt.Errorf("oplog many: wrapper is missing: %w", ErrNilParameter)
	}
	replayable, err := validateOplogArgs(items[0], opts)
	if err != nil {
		return fmt.Errorf("oplog many: oplog validation failed %w", err)
	}
	ticketer, err := oplog.NewGormTicketer(rw.underlying, oplog.WithAggregateNames(true))
	if err != nil {
		return fmt.Errorf("oplog many: unable to get Ticketer %w", err)
	}
	entry, err := oplog.NewEntry(
		replayable.TableName(),
		oplogArgs.metadata,
		oplogArgs.wrapper,
		ticketer,
	)
	if err != nil {
		return fmt.Errorf("oplog many: unable to create oplog entry %w", err)
	}
	oplogMsgs := []*oplog.Message{}
	var foundType reflect.Type
	for i, item := range items {
		if i == 0 {
			foundType = reflect.TypeOf(item)
		}
		currentType := reflect.TypeOf(item)
		if foundType != currentType {
			return fmt.Errorf("oplog many: items contains disparate types.  item %d is not a %s", i, foundType.Name())
		}
		replayable, ok := item.(oplog.ReplayableMessage)
		if !ok {
			return fmt.Errorf("oplog many: item %d not a replayable oplog message %w", i, ErrInvalidParameter)
		}
		msg := &oplog.Message{
			Message:  item.(proto.Message),
			TypeName: replayable.TableName(),
		}
		switch opType {
		case CreateOp:
			msg.OpType = oplog.OpType_OP_TYPE_CREATE
		case DeleteOp:
			msg.OpType = oplog.OpType_OP_TYPE_DELETE
		default:
			return fmt.Errorf("oplog many: operation type %v is not supported", opType)
		}
		oplogMsgs = append(oplogMsgs, msg)
	}
	if err := entry.WriteEntryWith(
		ctx,
		&oplog.GormWriter{Tx: rw.underlying},
		ticket,
		oplogMsgs...,
	); err != nil {
		return fmt.Errorf("oplog many: unable to write oplog entry %w", err)
	}
	return nil
}
func (rw *Db) addOplog(ctx context.Context, opType OpType, opts Options, ticket *store.Ticket, i interface{}) error {
	oplogArgs := opts.oplogOpts
	replayable, err := validateOplogArgs(i, opts)
	if err != nil {
		return err
	}
	if ticket == nil {
		return fmt.Errorf("add oplog: missing ticket %w", ErrNilParameter)
	}
	ticketer, err := oplog.NewGormTicketer(rw.underlying, oplog.WithAggregateNames(true))
	if err != nil {
		return fmt.Errorf("add oplog: unable to get Ticketer %w", err)
	}
	entry, err := oplog.NewEntry(
		replayable.TableName(),
		oplogArgs.metadata,
		oplogArgs.wrapper,
		ticketer,
	)
	if err != nil {
		return err
	}
	msg := oplog.Message{
		Message:  i.(proto.Message),
		TypeName: replayable.TableName(),
	}
	switch opType {
	case CreateOp:
		msg.OpType = oplog.OpType_OP_TYPE_CREATE
	case UpdateOp:
		msg.OpType = oplog.OpType_OP_TYPE_UPDATE
		msg.FieldMaskPaths = opts.WithFieldMaskPaths
		msg.SetToNullPaths = opts.WithNullPaths
	case DeleteOp:
		msg.OpType = oplog.OpType_OP_TYPE_DELETE
	default:
		return fmt.Errorf("add oplog: operation type %v is not supported", opType)
	}
	err = entry.WriteEntryWith(
		ctx,
		&oplog.GormWriter{Tx: rw.underlying},
		ticket,
		&msg,
	)
	if err != nil {
		return fmt.Errorf("add oplog: unable to write oplog entry: %w", err)
	}
	return nil
}

// DoTx will wrap the Handler func passed within a transaction with retries
// you should ensure that any objects written to the db in your TxHandler are retryable, which
// means that the object may be sent to the db several times (retried), so things like the primary key must
// be reset before retry
func (w *Db) DoTx(ctx context.Context, retries uint, backOff Backoff, Handler TxHandler) (RetryInfo, error) {
	if w.underlying == nil {
		return RetryInfo{}, errors.New("do underlying db is nil")
	}
	info := RetryInfo{}
	for attempts := uint(1); ; attempts++ {
		if attempts > retries+1 {
			return info, fmt.Errorf("Too many retries: %d of %d", attempts-1, retries+1)
		}

		// step one of this, start a transaction...
		newTx := w.underlying.BeginTx(ctx, nil)

		rw := &Db{newTx}
		if err := Handler(rw, rw); err != nil {
			if err := newTx.Rollback().Error; err != nil {
				return info, err
			}
			if errors.Is(err, oplog.ErrTicketAlreadyRedeemed) {
				d := backOff.Duration(attempts)
				info.Retries++
				info.Backoff = info.Backoff + d
				time.Sleep(d)
				continue
			}
			return info, err
		}

		if err := newTx.Commit().Error; err != nil {
			if err := newTx.Rollback().Error; err != nil {
				return info, err
			}
			return info, err
		}
		return info, nil // it all worked!!!
	}
}

// LookupByName will lookup resource my its friendly name which must be unique
func (rw *Db) LookupByName(ctx context.Context, resource ResourceNamer, opt ...Option) error {
	if rw.underlying == nil {
		return errors.New("error underlying db nil for lookup by name")
	}
	if reflect.ValueOf(resource).Kind() != reflect.Ptr {
		return errors.New("error interface parameter must to be a pointer for lookup by name")
	}
	if resource.GetName() == "" {
		return errors.New("error name empty string for lookup by name")
	}
	if err := rw.underlying.Where("name = ?", resource.GetName()).First(resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

// LookupByPublicId will lookup resource my its public_id which must be unique
func (rw *Db) LookupByPublicId(ctx context.Context, resource ResourcePublicIder, opt ...Option) error {
	if rw.underlying == nil {
		return errors.New("error underlying db nil for lookup by public id")
	}
	if reflect.ValueOf(resource).Kind() != reflect.Ptr {
		return errors.New("error interface parameter must to be a pointer for lookup by public id")
	}
	if resource.GetPublicId() == "" {
		return errors.New("error public id empty string for lookup by public id")
	}
	if err := rw.underlying.Where("public_id = ?", resource.GetPublicId()).First(resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

// LookupWhere will lookup the first resource using a where clause with parameters (it only returns the first one)
func (rw *Db) LookupWhere(ctx context.Context, resource interface{}, where string, args ...interface{}) error {
	if rw.underlying == nil {
		return errors.New("error underlying db nil for lookup by")
	}
	if reflect.ValueOf(resource).Kind() != reflect.Ptr {
		return errors.New("error interface parameter must to be a pointer for lookup by")
	}
	if err := rw.underlying.Where(where, args...).First(resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return ErrRecordNotFound
		}
		return err
	}
	return nil
}

// SearchWhere will search for all the resources it can find using a where
// clause with parameters.  Supports the WithLimit option.  If
// WithLimit < 0, then unlimited results are returned.  If WithLimit == 0, then
// default limits are used for results.
func (rw *Db) SearchWhere(ctx context.Context, resources interface{}, where string, args []interface{}, opt ...Option) error {
	opts := GetOpts(opt...)
	if rw.underlying == nil {
		return errors.New("error underlying db nil for search by")
	}
	if reflect.ValueOf(resources).Kind() != reflect.Ptr {
		return errors.New("error interface parameter must to be a pointer for search by")
	}
	var err error
	switch {
	case opts.WithLimit < 0: // any negative number signals unlimited results
		err = rw.underlying.Where(where, args...).Find(resources).Error
	case opts.WithLimit == 0: // zero signals the default value and default limits
		err = rw.underlying.Limit(DefaultLimit).Where(where, args...).Find(resources).Error
	default:
		err = rw.underlying.Limit(opts.WithLimit).Where(where, args...).Find(resources).Error
	}
	if err != nil {
		// searching with a slice parameter does not return a gorm.ErrRecordNotFound
		return err
	}
	return nil
}

// filterPaths will filter out non-updatable fields
func filterPaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	filtered := []string{}
	for _, p := range paths {
		switch {
		case strings.EqualFold(p, "CreateTime"):
			continue
		case strings.EqualFold(p, "UpdateTime"):
			continue
		case strings.EqualFold(p, "PublicId"):
			continue
		default:
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func setFieldsToNil(i interface{}, fieldNames []string) {
	if err := Clear(i, fieldNames, 2); err != nil {
		// do nothing
	}
}

func isNil(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func contains(ss []string, t string) bool {
	for _, s := range ss {
		if strings.EqualFold(s, t) {
			return true
		}
	}
	return false
}

// Clear sets fields in the value pointed to by i to their zero value.
// Clear descends i to depth clearing fields at each level. i must be a
// pointer to a struct. Cycles in i are not detected.
//
// A depth of 2 will change i and i's children. A depth of 1 will change i
// but no children of i. A depth of 0 will return with no changes to i.
func Clear(i interface{}, fields []string, depth int) error {
	if len(fields) == 0 || depth == 0 {
		return nil
	}
	fm := make(map[string]bool)
	for _, f := range fields {
		fm[f] = true
	}

	v := reflect.ValueOf(i)

	switch v.Kind() {
	default:
		return ErrInvalidParameter
	case reflect.Ptr:
		if v.IsNil() || v.Elem().Kind() != reflect.Struct {
			return ErrInvalidParameter
		}
		clear(v, fm, depth)
	}
	return nil
}

func clear(v reflect.Value, fields map[string]bool, depth int) {
	if depth == 0 {
		return
	}
	depth--

	switch v.Kind() {
	case reflect.Ptr:
		clear(v.Elem(), fields, depth+1)
	case reflect.Struct:
		typeOfT := v.Type()
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			if ok := fields[typeOfT.Field(i).Name]; ok {
				if f.IsValid() && f.CanSet() {
					f.Set(reflect.Zero(f.Type()))
				}
				continue
			}
			clear(f, fields, depth)
		}
	}
}
