package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"ebayclone/changeset"
	"github.com/go-sql-driver/mysql"
)

type DefaultConfigQuery struct {
	IncludeColAs  bool
	RenameTableAs string
}

var repo *Repo

func print(v string, debug bool) {
	if debug {
		fmt.Println(v)
	}
}

type Repo struct {
	db    *sql.DB
	debug bool
}

type ErrCode uint32

const (
	ErrCodeDuplicate ErrCode = iota + 1
	ErrCodeNotFoundParentKey
	ErrCodeNotFoundUpdateIdEntity
	OtherErrCode
)

var customPrefixUpdateNotFound = "Error Update Custom: Not Found Id"

func GetErrCode(myErr error) ErrCode {
	if strings.HasPrefix(myErr.Error(), "Error 1062") {
		return ErrCodeDuplicate
	}
	if strings.HasPrefix(myErr.Error(), "Error 1452") {
		return ErrCodeNotFoundParentKey
	}
	if strings.HasPrefix(myErr.Error(), customPrefixUpdateNotFound) {
		return ErrCodeNotFoundUpdateIdEntity
	}
	return OtherErrCode
}
func NewRepo(config *mysql.Config, debug bool) *Repo {
	if repo != nil {
		return repo
	}
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		//replace fmt.Println
		return nil
	}
	//replace fmt.Println
	//replace fmt.Println
	repo = &Repo{
		db:    db,
		debug: debug,
	}
	return repo
}

type Querier interface {
	query(config ...*DefaultConfigQuery) (query string, args []interface{})
	Append(Querier) Querier
}

type IFNULLTYPE uint8

const (
	IFNULLINT IFNULLTYPE = iota + 1
	IFNULLSTR
)

type C struct {
	name              string
	table             string
	as                string
	isDateType        bool
	dateTimeConverter *DateTimeConverter
	ifNullType        IFNULLTYPE
}

func Col(name string, tb string, IfNullReplaceZero ...IFNULLTYPE) *C {
	c := &C{
		name:  name,
		table: tb,
	}
	if len(IfNullReplaceZero) == 1 {
		c.ifNullType = IfNullReplaceZero[0]
	}
	return c
}

func (c *C) As(as string) *C {
	c.as = as
	return c
}

type Selector struct {
	cols []*C
}

func (s *Selector) Append(querier Querier) Querier {
	if _, ok := querier.(*Selector); ok {
		s.cols = append(s.cols, querier.(*Selector).cols...)
	}
	return s
}

func (s *Selector) query(config ...*DefaultConfigQuery) (string, []interface{}) {
	if len(s.cols) == 0 {
		return "", nil
	}

	q := ""
	for i, col := range s.cols {
		var tempRename string = col.table
		if len(config) > 0 {
			tempRename = config[0].RenameTableAs
			if !config[0].IncludeColAs {
				col.as = ""
			}
		}

		if col.dateTimeConverter != nil {
			col_converted := fmt.Sprintf("`%v`.`%v`", tempRename, col.name)
			date_select_query := ""
			if col.dateTimeConverter.TimeZoneConvertFormat != "" {
				date_select_query = fmt.Sprintf("CONVERT_TZ(%v, '+00:00', '%v')", col_converted, col.dateTimeConverter.TimeZoneConvertFormat)
			}
			if col.dateTimeConverter.DateFormat != "" {
				if date_select_query != "" {
					date_select_query = fmt.Sprintf("DATE_FORMAT(%v, '%v')", date_select_query, col.dateTimeConverter.DateFormat)
				} else {
					date_select_query = fmt.Sprintf("DATE_FORMAT(%v, '%v')", col_converted, col.dateTimeConverter.DateFormat)
				}
			}

			// reset col_converted
			col_converted = col.name
			if col.as != "" {

				col_converted = col.as
			}
			q += fmt.Sprintf("%v AS %v", date_select_query, col_converted)
		} else {
			if col.ifNullType > 0 {
				if col.ifNullType == IFNULLINT {
					q += fmt.Sprintf("IFNULL(`%v`.`%v`, 0)", tempRename, col.name)
				}
				if col.ifNullType == IFNULLSTR {
					q += fmt.Sprintf("IFNULL(`%v`.`%v`, '')", tempRename, col.name)
				}
			} else {
				q += fmt.Sprintf("`%v`.`%v`", tempRename, col.name)
			}
		}
		if col.as != "" && col.dateTimeConverter == nil {
			q += " AS "
			q += col.as
		}
		if i < len(s.cols)-1 {
			q += ", "
		}
	}
	q += " "
	return q, nil
}

type QueryBuilder struct {
	query      string
	table      string
	Projection Querier
	Predicate  Querier
	limit      Querier
	groupBy    Querier
	orderBy    Querier
	args       []interface{}
}

func (q *QueryBuilder) OrderBy(c *C, orderType OrderType) *QueryBuilder {
	o := &orderBy{
		name:      c.name,
		table:     c.table,
		orderType: orderType,
	}
	q.orderBy = o
	return q
}

func (q *QueryBuilder) Query() (string, []interface{}) {
	query := ""
	if q.query != "" {
		query = q.query
	}
	if q.Projection != nil {
		projectQuery, args := q.Projection.query()
		if args != nil {
			q.args = append(q.args, args)
		}

		query = " SELECT " + projectQuery
		query += q.query
		q.query = query
	}
	if q.Predicate != nil {
		q.query += " "
		predicateQuery, args := q.Predicate.query()
		q.query += predicateQuery
		if args != nil {
			q.args = append(q.args, args...)
		}
	}
	if q.orderBy != nil {
		q.query += " "
		orderByQuery, _ := q.orderBy.query()
		q.query += orderByQuery
	}
	return q.query, q.args
}

func (q *QueryBuilder) Select(col *C) *QueryBuilder {
	if q.Projection == nil {
		q.Projection = &Selector{
			cols: make([]*C, 0),
		}
	}
	q.Projection.(*Selector).cols = append(q.Projection.(*Selector).cols, col)
	return q
}

type DateTimeConverter struct {
	TimeZoneConvertFormat string
	DateFormat            string
}

func NewDateTimeConverter(timeZoneFormat string, DateFormat string) *DateTimeConverter {
	return &DateTimeConverter{
		TimeZoneConvertFormat: timeZoneFormat,
		DateFormat:            DateFormat,
	}
}
func (q *QueryBuilder) SelectDate(col *C, dateTimeConverter ...*DateTimeConverter) *QueryBuilder {
	col.isDateType = true
	if len(dateTimeConverter) > 0 {
		col.dateTimeConverter = dateTimeConverter[0]
	}
	q.Select(col)
	return q
}

type PredicateOp uint

func (p PredicateOp) toString() string {
	if p == LessEqual {
		return "<="
	}
	if p == Less {
		return "<"
	}
	if p == GreaterEqual {
		return ">="
	}

	if p == Greater {
		return ">"
	}

	if p == Like {
		return "LIKE"
	}

	if p == NotEqual {
		return "!="
	}
	if p == ISNULL {
		return "IS NULL"
	}
	if p == ISNOTNULL {
		return "IS NOT NULL"
	}
	return "="
}

const (
	LessEqual PredicateOp = iota
	Less
	GreaterEqual
	Greater
	Equal
	Like
	NotEqual
	In
	ISNULL
	ISNOTNULL
)

type PrefixOp uint8

func (o *PrefixOp) ToOpString() string {
	if (*o) == PrefixOr {
		return "OR"
	}
	return "AND"
}

const (
	PrefixAnd PrefixOp = iota
	PrefixOr
)

type Predicate struct {
	prefixOp PrefixOp
	col      string
	table    string
	op       string
	val      interface{}
	depth    int
	down     *Predicate
}

func OrP(col string, table string, op PredicateOp, val ...interface{}) *Predicate {
	p := &Predicate{
		prefixOp: PrefixOr,
		col:      col,
		table:    table,
		op:       op.toString(),
		depth:    0,
	}
	if len(val) == 1 {
		p.val = val[0]
	}
	return p
}

func P(col string, table string, op PredicateOp, val ...interface{}) *Predicate {
	p := &Predicate{
		prefixOp: PrefixAnd,
		col:      col,
		table:    table,
		op:       op.toString(),
		depth:    0,
	}
	if len(val) == 1 {
		p.val = val[0]
	}
	return p
}

func Or(predicates ...*Predicate) *Predicate {
	if len(predicates) == 1 {
		return predicates[0]
	}
	f := predicates[0]
	n := len(predicates) - 1
	if f.down == nil {
		f = predicates[len(predicates)-1]
		n--
	}
	d := f
	is_block := 0
	for d != nil {
		if d.down != nil {
			is_block = 1
		}
		if is_block == 1 {
			d.depth++
		}
		d = d.down
	}
	m := n
	v := predicates[n]
	for m > 0 {
		m--
		if predicates[m] != f {
			v.down = predicates[m]
			v = v.down
		}
	}
	v.down = f
	//k := predicates[n]
	////replace fmt.Println
	predicates[n].prefixOp = PrefixOr
	return predicates[n]
}

func And(predicates ...*Predicate) *Predicate {
	if len(predicates) == 1 {
		return predicates[0]
	}
	f := predicates[0]
	n := len(predicates) - 1
	if f.down == nil {
		f = predicates[len(predicates)-1]
		n--
	}
	d := f
	is_block := 0
	for d != nil {
		if d.down != nil {
			is_block = 1
		}
		if is_block == 1 {
			d.depth++
		}
		d = d.down
	}
	m := n
	v := predicates[n]
	for m > 0 {
		m--
		if predicates[m] != f {
			v.down = predicates[m]
			v = v.down
		}
	}
	v.down = f
	//k := predicates[n]
	////replace fmt.Println
	predicates[n].prefixOp = PrefixAnd
	return predicates[n]
}

type Where struct {
	curBlock   int
	predicates []*Predicate
}

func (w *Where) Append(querier Querier) Querier {
	if _, check := querier.(*Where); check {
		w.predicates = append(w.predicates, querier.(*Where).predicates...)
	}
	return w
}

func (w *Where) query(config ...*DefaultConfigQuery) (string, []interface{}) {
	query := "WHERE "
	arguments := []interface{}{}
	p := w.predicates[0]
	curDepth := 0
	//first := true
	num_ends := 0
	for p != nil {
		// if depth > 0 and depth not equal curDepth
		// mean we have one depth nested within
		// then open "(" will created
		// also mark be added to 1, to can add end ")" later
		if p.depth > 0 && p.depth != curDepth {
			query += "("
			curDepth = p.depth
			num_ends++
			//first = false
		}
		query += fmt.Sprintf("`%v`.`%v` %v ", p.table, p.col, p.op)
		if p.op != "IS NULL" && p.op != "IS NOT NULL" {
			query += "?"
			arguments = append(arguments, p.val)
		}

		if p.down != nil {
			query += " " + p.prefixOp.ToOpString() + " "
		}
		p = p.down
	}
	for num_ends > 0 {
		query += ")"
		num_ends--
	}
	//for i, p := range w.predicates {
	//	if i > 0 {
	//		if p.prefixOp == PrefixAnd {
	//			query += " AND "
	//		} else {
	//			query += " OR "
	//		}
	//	}
	//	if len(config) > 0 {
	//		p.table = config[0].RenameTableAs
	//	}
	//	query += fmt.Sprintf("`%v`.`%v` %v ", p.table, p.col, p.op)
	//	if p.op != "IS NULL" {
	//		query += "?"
	//		arguments = append(arguments, p.val)
	//	}
	//
	//}
	return query, arguments
}

func (q *QueryBuilder) Where(predicate *Predicate) *QueryBuilder {
	if q.Predicate == nil {
		q.Predicate = &Where{
			predicates: make([]*Predicate, 0),
		}
		q.Predicate.(*Where).predicates = append(q.Predicate.(*Where).predicates, predicate)
		return q
	}
	p := q.Predicate.(*Where).predicates[0]
	for p != nil {
		if p.down == nil {
			break
		}
		p = p.down
	}
	p.down = predicate
	return q
}

func (q *QueryBuilder) Wheres(predicate *Predicate) *QueryBuilder {
	if q.Predicate == nil {
		q.Predicate = &Where{
			predicates: make([]*Predicate, 0),
		}
		q.Predicate.(*Where).predicates = append(q.Predicate.(*Where).predicates, predicate)
		return q
	}
	p := q.Predicate.(*Where).predicates[0]
	for p != nil {
		if p.down == nil {
			break
		}
		p = p.down
	}
	d := predicate
	is_block := 0
	for d != nil {
		if d.down != nil {
			is_block = 1
		}
		if is_block == 1 {
			d.depth++
		}
		d = d.down
	}
	p.down = predicate

	return q
}

type TYPEJOIN uint8

const (
	LEFTJOIN TYPEJOIN = iota + 1
	RIGHTJOIN
	INNERJOIN
)

func (t TYPEJOIN) ToQueryString() string {
	if t == LEFTJOIN {
		return "LEFT JOIN"
	}
	if t == RIGHTJOIN {
		return "RIGHT JOIN"
	}
	return "INNER JOIN"
}
func (r *Repo) GetById(need interface{}, preloads ...func() (to interface{}, fk string, pk string, inverse bool, type_join TYPEJOIN)) *QueryBuilder {
	nv := reflect.Indirect(reflect.ValueOf(need))
	nvName := nv.Type().Name()
	nvTable := strings.ToLower(nvName) + "s"
	if len(preloads) == 0 {
		return &QueryBuilder{
			query: fmt.Sprintf("FROM `%v`", nvTable),
			table: nvTable,
			args:  []interface{}{},
		}
	}
	if len(preloads) == 1 {
		to, fk, pk, inverse, typ := preloads[0]()
		pv := reflect.Indirect(reflect.ValueOf(to))
		//replace fmt.Println
		pvName := pv.Type().Name()

		nvKey := pk
		pvTable := strings.ToLower(pvName) + "s"
		pvKey := fk
		if inverse {
			nvKey, pvKey = pvKey, nvKey
		}
		typ_join := typ.ToQueryString()
		query := fmt.Sprintf("FROM `%v` %v `%v` ON `%v`.`%v` = `%v`.`%v`", nvTable, typ_join, pvTable, nvTable, nvKey, pvTable, pvKey)
		return &QueryBuilder{table: nvTable, query: query, args: []interface{}{}}
	}
	// multiple joiners
	query := ""
	for i, preload := range preloads {
		to, fk, pk, inverse, typ := preload()
		pv := reflect.Indirect(reflect.ValueOf(to))
		//replace fmt.Println
		pvName := pv.Type().Name()

		nvKey := pk
		pvTable := strings.ToLower(pvName) + "s"
		pvKey := fk
		if inverse {
			nvKey, pvKey = pvKey, nvKey
		}
		if i == 0 {
			query += fmt.Sprintf("FROM `%v`", nvTable)
		} else {
			query += " "
		}
		typ_join := typ.ToQueryString()
		query += fmt.Sprintf("%v `%v` ON `%v`.`%v` = `%v`.`%v`", typ_join, pvTable, nvTable, nvKey, pvTable, pvKey)
	}
	return &QueryBuilder{table: nvTable, query: query, args: []interface{}{}}
}

type Condition struct {
	OrderBy bool
}

// update repo
type RelList struct {
	Level []interface{}
}

type RelRelation struct {
	relScaned *reflect.Value
	fieldRef  string
	isO2O     bool
}

func (r *Repo) ParseToStruct(rows *sql.Rows, cast interface{}, cond ...*Condition) ([]interface{}, []interface{}) {
	cols, err := rows.Columns()
	if err != nil {
		return nil, nil
	}
	var scaned = map[interface{}]reflect.Value{}
	var orderId = []interface{}{}
	//replace fmt.Println
	castReflect := reflect.Indirect(reflect.ValueOf(cast))
	results := []interface{}{}
	count := 0
	for rows.Next() {
		count += 1
		jsonByteAddrs := map[string][]interface{}{}
		castedNew := reflect.Indirect(reflect.New(castReflect.Type()))
		print(fmt.Sprintf("casted new: %v\n", castedNew.Type()), r.debug)
		addrs := make([]interface{}, len(cols))
		rels := map[string]*RelRelation{}
		for i, col := range cols {
			_, ok := castedNew.Type().FieldByName(col)
			isJson := false
			if !ok {
				str := strings.Split(col, "$")
				scName := strings.Split(str[0], "Rel")
				if _, jsonOk := changeset.JsonFieldsOfSchemas[scName[0]]; jsonOk {
					if _, jsonOk := changeset.JsonFieldsOfSchemas[scName[0]][str[1]]; jsonOk {
						isJson = true
					}
				}
				////replace fmt.Println
				if len(str) != 2 {
					var addr interface{}
					addrs[i] = &addr
					continue
				}
				if _, ok := rels[str[0]]; !ok {
					rels[str[0]] = &RelRelation{
						isO2O: false,
					}
				}
				rel := rels[str[0]]
				fRel := castedNew.FieldByName(str[0])
				var fRelType reflect.Type
				//fmt.Println("before panic: ", fRel, castedNew.FieldByName(str[0]), " field name: ", str[0])
				if fRel.Type().Kind() == reflect.Slice {
					fRelType = fRel.Type().Elem()
				}
				if fRel.Type().Kind() == reflect.Ptr {
					fRelType = fRel.Type()
					rel.isO2O = true
				}
				// fRelType is pointer => use elem before use new
				rel.fieldRef = str[0]
				if rel.relScaned == nil {
					var newRel = reflect.New(fRelType.Elem())
					rel.relScaned = &newRel
				}
				_, ok = (*rel.relScaned).Type().Elem().FieldByName(str[1])
				if isJson {
					var addr []byte
					if _, ok := jsonByteAddrs[str[1]]; !ok {
						jsonByteAddrs[str[1]] = make([]interface{}, 2)
					}
					//replace fmt.Println
					jsonByteAddrs[str[1]][0] = &addr
					jsonByteAddrs[str[1]][1] = rel.relScaned
					//replace fmt.Println
					addrs[i] = &addr
					continue
				}
				if !ok {
					// check field of rel name exist if not assign empty addr
					var addr interface{}
					addrs[i] = &addr
					continue
				}
				if _, ok := (*rel.relScaned).Elem().Type().FieldByName(str[1]); ok {
					//replace fmt.Println
				}
				addrs[i] = (*rel.relScaned).Elem().FieldByName(str[1]).Addr().Interface()
				continue
			}
			scName := castReflect.Type().Name()
			if _, jsonOk := changeset.JsonFieldsOfSchemas[scName]; jsonOk {
				if _, jsonOk := changeset.JsonFieldsOfSchemas[scName][col]; jsonOk {
					isJson = true
				}
			}
			if isJson {
				var addr []byte
				if _, ok := jsonByteAddrs[col]; !ok {
					jsonByteAddrs[col] = make([]interface{}, 2)
				}
				//fmt.Println("json field: ", col)
				jsonByteAddrs[col][0] = &addr
				jsonByteAddrs[col][1] = &castedNew
				//replace fmt.Println
				addrs[i] = &addr
				continue
			}

			f := castedNew.FieldByName(col)
			//replace fmt.Println
			addrs[i] = f.Addr().Interface()
		}
		rows.Scan(addrs...)
		for fieldName, byteAndStructAddrJson := range jsonByteAddrs {

			if byteAddr, ok := byteAndStructAddrJson[0].(*[]byte); ok {
				var reflectTypeJsonClass reflect.Type
				var havePointerSet bool = false
				if byteAndStructAddrJson[1].(*reflect.Value).Kind() == reflect.Ptr {
					reflectTypeJsonClass = byteAndStructAddrJson[1].(*reflect.Value).Elem().FieldByName(fieldName).Type().Elem()
					havePointerSet = true
				}
				if byteAndStructAddrJson[1].(*reflect.Value).Kind() == reflect.Struct {
					reflectTypeJsonClass = byteAndStructAddrJson[1].(*reflect.Value).FieldByName(fieldName).Type().Elem()
				}
				rvClassJson := reflect.New(reflectTypeJsonClass).Interface()
				if err := json.Unmarshal(*byteAddr, rvClassJson); err == nil {
					//replace fmt.Println
					//fmt.Println("cast json success: ", rvClassJson, string(*byteAddr))
					if havePointerSet {
						byteAndStructAddrJson[1].(*reflect.Value).Elem().FieldByName(fieldName).Set(reflect.ValueOf(rvClassJson))
					} else {
						byteAndStructAddrJson[1].(*reflect.Value).FieldByName(fieldName).Set(reflect.ValueOf(rvClassJson))
					}
				} else {
					//fmt.Println("cast json error: ", err)
					//replace fmt.Println
				}
			}
		}
		if _, ok := castedNew.Type().FieldByName("Id"); ok {
			idVal := castedNew.FieldByName("Id").Interface()
			if _, ok := scaned[idVal]; !ok {
				scaned[idVal] = castedNew
				if len(cond) == 1 {
					if cond[0].OrderBy {
						orderId = append(orderId, idVal) // sort id of order by
					}
				}
				//replace fmt.Println
			}
			//replace fmt.Println
			for _, rel := range rels {
				if !rel.isO2O {
					newVal := reflect.Append(scaned[idVal].FieldByName(rel.fieldRef), *rel.relScaned)
					scaned[idVal].FieldByName(rel.fieldRef).Set(newVal)
				} else {
					if scaned[idVal].FieldByName(rel.fieldRef).CanSet() {
						scaned[idVal].FieldByName(rel.fieldRef).Set(*rel.relScaned)
					} else {
						//replace fmt.Println
					}
				}
			}
			//for _, addr := range addrs {
			//	rv := reflect.Indirect(reflect.ValueOf(addr))
			//	//replace fmt.Println
			//}
		}
	}
	print(fmt.Sprintf("TOTAL ROW SCANNED FROM MYSQL: %v\n", count), r.debug)
	if len(orderId) == 0 {
		for k, v := range scaned {
			results = append(results, v.Addr().Interface())
			delete(scaned, k)
		}
		return results, nil
	}
	for _, id := range orderId {
		//replace fmt.Println
		if _, ok := scaned[id]; ok {
			results = append(results, scaned[id].Addr().Interface())
			delete(scaned, id)
		}
	}
	return results, nil
}

func (r *Repo) RawQuery(query string, args []interface{}, cast interface{}) ([]interface{}, []interface{}) {
	stmt, err := r.db.Prepare(query)
	if err != nil {
		//replace fmt.Println
		fmt.Println("[Log-RawQuery], prepare statement error: ", err)
		return nil, nil
	}
	rows, err := stmt.Query(args...)
	if err != nil {
		//replace fmt.Println
		print(fmt.Sprintf("[Log-RawQuery], error: %v\n", err), r.debug)
	}

	//replace fmt.Println
	if strings.Contains(query, "ORDER BY") {
		return r.ParseToStruct(rows, cast, &Condition{OrderBy: true})
	}
	return r.ParseToStruct(rows, cast)
}

func (r *Repo) Save(ctx context.Context, cs *changeset.ChangeSet) error {
	query, args := r.insertQuery(cs)
	stmt, err := r.db.PrepareContext(ctx, query)
	//replace fmt.Println
	if err != nil {
		//replace fmt.Println
		return err
	}
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		//replace fmt.Println
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	cs.ReflectSchema.FieldByName("Id").Set(reflect.ValueOf(uint32(id)))
	cs.ActionRepo = changeset.ActionInsert
	return nil
}

func (r *Repo) SaveTx(ctx context.Context, cs *changeset.ChangeSet, tx *sql.Tx) error {
	query, args := r.insertQuery(cs)
	stmt, err := tx.PrepareContext(ctx, query)
	//replace fmt.Println
	if err != nil {
		//replace fmt.Println
		return err
	}
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		//replace fmt.Println
		return err
	}
	id, err := result.LastInsertId()
	if err != nil {
		//replace fmt.Println
		return err
	}

	if (cs.Boxes["Id"].GetOps() & (1 << changeset.AI)) != 0 {
		cs.ReflectSchema.FieldByName("Id").Set(reflect.ValueOf(uint32(id)))
	}
	cs.ActionRepo = changeset.ActionInsert
	return nil
}

func (r *Repo) OpenTx(ctx context.Context) *sql.Tx {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
	if err != nil {
		return nil
	}
	return tx
}

func (r *Repo) insertQuery(cs *changeset.ChangeSet) (string, []interface{}) {
	tb := strings.ToLower(cs.ReflectSchema.Type().Name()) + "s"
	query := fmt.Sprintf("INSERT INTO `%v` (", tb)
	values := " VALUES ("
	args := []interface{}{}
	for i, col := range cs.CastedBoxes {
		if cs.Boxes[col].UpdatedCol != "" {
			query += fmt.Sprintf("`%v`", cs.Boxes[col].RelTbName+cs.Boxes[col].UpdatedCol)
		} else {
			query += fmt.Sprintf("`%v`", col)
		}
		values += "?"
		if i < len(cs.CastedBoxes)-1 {
			query += ", "
			values += ", "
		}
		args = append(args, cs.Boxes[col].GetVal())
	}
	query += ")"
	values += ")"
	query += values
	return query, args
}

func (r *Repo) UpdateById(ctx context.Context, cs *changeset.ChangeSet) error {
	query, args := UpdateQuery(cs)
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		//replace fmt.Println
		return err
	}

	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		//replace fmt.Println
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		//fmt.Printf("[REPO_UPDATE]: error, [err=%v]\n", err)
		return fmt.Errorf(customPrefixUpdateNotFound)
	}
	if n < 1 {
		//fmt.Printf("[REPO_UPDATE]: error, [err=%v]\n", err)
		return fmt.Errorf(customPrefixUpdateNotFound)
	}
	cs.ActionRepo = changeset.ActionUpdate
	return nil
}

func (r *Repo) UpdateTxById(ctx context.Context, cs *changeset.ChangeSet, tx *sql.Tx, append_query ...string) error {
	query, args := UpdateQuery(cs, append_query...)
	fmt.Println("update by transaction query: ", query, args)
	//replace fmt.Println
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		//replace fmt.Println
		return err
	}
	result, err := stmt.ExecContext(ctx, args...)
	if err != nil {
		//replace fmt.Println
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(customPrefixUpdateNotFound)
	}
	if n < 1 {
		return fmt.Errorf(customPrefixUpdateNotFound)
	}
	cs.ActionRepo = changeset.ActionUpdate
	return nil
}

func UpdateQuery(cs *changeset.ChangeSet, append_query ...string) (string, []interface{}) {
	tbName := strings.ToLower(cs.ReflectSchema.Type().Name()) + "s"
	query := fmt.Sprintf("UPDATE %v SET ", tbName)
	args := []interface{}{}
	for i, col := range cs.CastedBoxes {
		have_nil := false
		if cs.Boxes[col].UpdatedCol != "" {
			query += fmt.Sprintf("`%v` = ", cs.Boxes[col].RelTbName+cs.Boxes[col].UpdatedCol)
			rvVal := reflect.Indirect(reflect.ValueOf(cs.Boxes[col].GetVal()))
			if rvVal.IsZero() {
				query += " null"
				have_nil = true
			} else {
				query += " ?"
			}
		} else {
			query += fmt.Sprintf("`%v` = ?", col)
		}
		if !have_nil {
			args = append(args, cs.Boxes[col].GetVal())
		}
		if i < len(cs.CastedBoxes)-1 {
			query += ", "
		}
	}
	query += " WHERE `Id` = ?"
	if len(append_query) > 0 {
		query += append_query[0]
	}
	args = append(args, cs.ReflectSchema.FieldByName("Id").Interface())
	//replace fmt.Println
	return query, args
}

type Rel struct {
	from    string
	to      string
	fromKey string
	toKey   string
	builder *QueryBuilder
}

type CacheKey struct {
	key string
	tb  string
}

type QueryRel struct {
	rels            []*Rel
	query           string
	args            []interface{}
	index           int
	joinedKeysCache []*CacheKey
}

func (q *QueryRel) OpenRel(rel *Rel) *QueryRel {
	q.rels = append(q.rels, rel)
	return q
}

func (q *QueryRel) ParseToQuery() (string, []interface{}) {
	dfs(q, true)
	return q.query, q.args
}

func dfs(q *QueryRel, rebuild bool) {
	if rebuild {
		if q.index >= len(q.rels) {
			q.index = 0
			dfs(q, false)
			return
		}
		index := q.index
		source := q.rels[index]

		if index < len(q.rels) && index > 0 {
			if q.rels[index-1].from == source.from {
				swapFromAndTo(q.rels[index-1], index, len(q.rels))
			}
			if q.rels[index-1].from == source.to {
				swapFromAndTo(q.rels[index-1], index, len(q.rels))
			}
		}
		q.index++
		dfs(q, true)
	} else {
		if q.index >= len(q.rels) {
			return
		}
		index := q.index
		source := q.rels[index]
		var haveExpandQuery bool
		if index > 0 {
			q.query += "("
		}
		if index < len(q.rels)-1 {
			var IncludeColAs bool = false
			if index == 0 {
				IncludeColAs = true
			}
			for i := index; i < len(q.rels)-1; i++ {
				if q.rels[i+1].builder != nil {
					//replace fmt.Println
					projectQuery, _ := q.rels[i+1].builder.Projection.query(&DefaultConfigQuery{
						IncludeColAs:  IncludeColAs,
						RenameTableAs: fmt.Sprintf("%v_%v", "r", index+1),
					})
					if projectQuery != "" {
						if !haveExpandQuery {
							haveExpandQuery = true
							q.query += "SELECT "
						} else {
							q.query += ", "
						}
						q.query += projectQuery
					}
				}
			}
		}
		// load self builder
		if q.rels[index].builder != nil {
			if q.rels[index].builder.Projection != nil {
				selfProjectQuery, _ := q.rels[index].builder.Projection.query()
				if selfProjectQuery != "" {
					if !haveExpandQuery {
						haveExpandQuery = true
						q.query += "SELECT "
					} else {
						q.query += ", "
					}
					q.query += selfProjectQuery
				}
			}

		}
		if index > 0 {
			if haveExpandQuery {
				q.query += fmt.Sprintf(", `%v`.`%v`", q.rels[index-1].to, q.rels[index-1].toKey)
			} else {
				q.query += fmt.Sprintf("SELECT `%v`.`%v`", q.rels[index-1].to, q.rels[index-1].toKey)
			}
			q.query += " "
		}
		q.query += fmt.Sprintf("FROM %v INNER JOIN ", source.from)
		q.index++
		dfs(q, false)
		// backtracking
		if index == len(q.rels)-1 {
			q.query += fmt.Sprintf("%v ON `%v`.`%v` = `%v`.`%v`", source.to, source.from, source.fromKey, source.to, source.toKey)
		}
		if index > 0 {
			tbNameAs := fmt.Sprintf("%v_%v", "r", index)
			q.query += fmt.Sprintf(") AS `%v` ON `%v`.`%v` = `%v`.`%v`", tbNameAs, q.rels[index-1].from, q.rels[index-1].fromKey, tbNameAs, q.rels[index-1].toKey)
		}
		if q.rels[index].builder != nil {
			if q.rels[index].builder.Predicate != nil {
				selfPredicateQuery, args := q.rels[index].builder.Predicate.query()
				if selfPredicateQuery != "" {
					q.query += selfPredicateQuery
					q.args = append(q.args, args...)
				}
			}
		}
	}
}

func JoinMultipleBuilder(left, right *QueryBuilder) {
	lv := reflect.Indirect(reflect.ValueOf(left))
	rv := reflect.Indirect(reflect.ValueOf(right))
	for i := 0; i < lv.NumField(); i++ {
		if lv.Field(i).Type().Name() == "Querier" && lv.Type().Field(i).IsExported() {
			append_name := lv.Type().Field(i).Name
			if _, inLeft := lv.Type().FieldByName(append_name); inLeft {
				if _, inRight := rv.Type().FieldByName(append_name); inRight {
					//replace fmt.Println
					var from Querier
					var to Querier
					leftAppendVal := lv.FieldByName(append_name).Interface()
					rightAppendVal := rv.FieldByName(append_name).Interface()
					if leftAppendVal != nil {
						from = leftAppendVal.(Querier)
						if rightAppendVal != nil {
							to = rightAppendVal.(Querier)
						}
					} else {
						from = rightAppendVal.(Querier)
						if leftAppendVal != nil {
							to = leftAppendVal.(Querier)
						}
					}

					if from != nil {
						if to != nil {
							from.Append(to)
						}
						lv.FieldByName(append_name).Set(reflect.ValueOf(from))
					}
				}
			}
		}
	}
}

func swapFromAndTo(rel *Rel, indexRel int, n int) {
	//replace fmt.Println
	rel.fromKey, rel.toKey = rel.toKey, rel.fromKey
	if indexRel == n-1 {
		rel.from, rel.to = rel.to, rel.from
	} else {
		rel.from = rel.to
		rel.to = fmt.Sprintf("r_%v", indexRel+1)
	}
	//replace fmt.Println
}

func JoinProjectBuilder(query *string, projectQuery *string) {
	*query = *projectQuery + *query
}

func ReplaceStringHaveAs(query *string) {
	var startPoint int = 0
	var index = 0
	for index < len(*query) {
		//replace fmt.Println
		if string((*query)[index]) == "A" && string((*query)[index-1]) == " " && index > 0 {
			if index < len(*query)-3 {
				x, y := (*query)[index+1], (*query)[index+2]
				if string(x) == "S" && string(y) == " " {
					// z is start point of rename table
					startPoint = index - 1
				}
			}
		}
		if string((*query)[index]) == "," && startPoint != 0 {
			c := (*query)[:startPoint]
			d := (*query)[index:]
			*query = c
			*query += d
			index = startPoint
			startPoint = 0
			continue
		}
		index++
	}
}

func (r *Repo) GetCursorDB() *sql.DB {
	return r.db
}

func (b *QueryBuilder) SelectDateFormat(c *C, format string) *Selector {
	return &Selector{}
}

func (q *QueryBuilder) ClearWhere() {
	q.Predicate = nil
}

func (q *QueryBuilder) CloneBuilder() *QueryBuilder {
	return &QueryBuilder{
		query:      q.query,
		table:      q.table,
		Projection: q.Projection,
		Predicate:  q.Predicate,
		limit:      q.limit,
		groupBy:    q.groupBy,
		orderBy:    q.orderBy,
		args:       q.args,
	}
}

func DeleteQuery(cs *changeset.ChangeSet) (string, []interface{}) {
	tbName := strings.ToLower(cs.ReflectSchema.Type().Name()) + "s"
	query := fmt.Sprintf("DELETE FROM `%v` ", tbName)
	args := []interface{}{}
	query += "WHERE `Id` = ?"
	args = append(args, cs.ReflectSchema.FieldByName("Id").Interface())
	return query, args
}
func (r *Repo) DeleteUserById(ctx context.Context, changeset *changeset.ChangeSet) error {
	query, args := DeleteQuery(changeset)
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	result, err := stmt.ExecContext(ctx, args...)
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	//replace fmt.Println
	if n == 1 {
		return nil
	}
	return fmt.Errorf("%v", NotFoundErr)
}

var NotFoundErr = "NotFound Error Entity"
