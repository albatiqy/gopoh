package repository

type AttrFilter struct {
	Attr string
	Type string
	Val  string
}

type AttrOrder struct {
	Attr string
	Type string
}

type CursorData struct {
	NextToken string `json:"next_token"`
	PrevToken string `json:"prev_token"`
}

type FinderOptionCursor struct {
	Search             string
	PageSize           uint8
	CursorToken        string
	Orders             []AttrOrder
	Filters            []AttrFilter
	allowedOrderAttrs  []string
	allowedFilterAttrs []string
}

func (qOpt *FinderOptionCursor) AppendFilter(attr string, filterType string, val string) {
	qOpt.Filters = append(qOpt.Filters, AttrFilter{
		Attr: attr,
		Type: filterType,
		Val:  val,
	})
}

func (qOpt *FinderOptionCursor) AppendOrder(attr string, orderType string) {
	qOpt.Orders = append(qOpt.Orders, AttrOrder{
		Attr: attr,
		Type: orderType,
	})
}

func (qOpt *FinderOptionCursor) AllowedOrderAttrs(allowedOrderAttrs ...string) {
	qOpt.allowedOrderAttrs = allowedOrderAttrs
}

func (qOpt *FinderOptionCursor) AllowedFilterAttrs(allowedFilterAttrs ...string) {
	qOpt.allowedFilterAttrs = allowedFilterAttrs
}