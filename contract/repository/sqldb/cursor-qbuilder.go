package sqldb

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/albatiqy/gopoh/contract/repository"
)

type FinderCursorQueryBuilder struct {
	driverSpec         DriverSpec
	baseQuery          string
	whereRaws          []string
	finderOptionCursor repository.FinderOptionCursor
	colsMap            *ColsMap
	isPrevNav          bool
	cursorID           string
}

func (qBuilder *FinderCursorQueryBuilder) AddWhereRaw(strWhere string) *FinderCursorQueryBuilder {
	qBuilder.whereRaws = append(qBuilder.whereRaws, strWhere)
	return qBuilder
}

func (qBuilder *FinderCursorQueryBuilder) Build() (string, []interface{}, error) {
	if qBuilder.finderOptionCursor.CursorToken != "" {
		b, err := base64.RawStdEncoding.DecodeString(qBuilder.finderOptionCursor.CursorToken)
		if err != nil {
			return "", nil, err
		}
		decodedToken := string(b)
		colonPos := strings.IndexRune(decodedToken, ':')
		if colonPos == -1 {
			return "", nil, fmt.Errorf("invalid cursor token")
		}
		if decodedToken[:colonPos] == "before" {
			qBuilder.isPrevNav = true
		}
		qBuilder.cursorID = decodedToken[colonPos+1:] //reset ??
	}
	if qBuilder.finderOptionCursor.PageSize == 0 {
		qBuilder.finderOptionCursor.PageSize = 4 // default pageSize
	}
	return qBuilder.driverSpec.BuildFinderCursorQuery(qBuilder.cursorID, qBuilder.isPrevNav, qBuilder.baseQuery, qBuilder.finderOptionCursor, qBuilder.whereRaws, qBuilder.colsMap)
}

func (qBuilder FinderCursorQueryBuilder) FillCursorData(cursorData *repository.CursorData, recordsLen int, getID func(itemsLen int, hasNext, isPrevNav bool) (string, string)) {
	if recordsLen > 0 {
		hasNext := (uint8(recordsLen) > qBuilder.finderOptionCursor.PageSize)
		firstID, lastID := getID(recordsLen-1, hasNext, qBuilder.isPrevNav)
		if qBuilder.isPrevNav {
			cursorData.NextToken = base64.RawStdEncoding.EncodeToString([]byte("after:" + lastID))
			if hasNext {
				cursorData.PrevToken = base64.RawStdEncoding.EncodeToString([]byte("before:" + firstID))
			}
		} else {
			if qBuilder.cursorID != "" {
				cursorData.PrevToken = base64.RawStdEncoding.EncodeToString([]byte("before:" + firstID))
			}
			if hasNext {
				cursorData.NextToken = base64.RawStdEncoding.EncodeToString([]byte("after:" + lastID))
			}
		}
	}
}
