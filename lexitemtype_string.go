// Code generated by "stringer -type=lexItemType"; DO NOT EDIT.

package sqlread

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TIllegal-0]
	_ = x[TEof-1]
	_ = x[TSemi-2]
	_ = x[TComma-3]
	_ = x[TComment-4]
	_ = x[TDelim-5]
	_ = x[TNull-6]
	_ = x[TString-7]
	_ = x[TNumber-8]
	_ = x[TIdentifier-9]
	_ = x[TDropTableFullStmt-10]
	_ = x[TLockTableFullStmt-11]
	_ = x[TUnlockTablesFullStmt-12]
	_ = x[TSetFullStmt-13]
	_ = x[TLParen-14]
	_ = x[TRParen-15]
	_ = x[TCreateTable-16]
	_ = x[TCreateTableDetail-17]
	_ = x[TCreateTableExtraDetail-18]
	_ = x[TColumnType-19]
	_ = x[TColumnSize-20]
	_ = x[TColumnEnumVal-21]
	_ = x[TColumnDetails-22]
	_ = x[TInsertInto-23]
	_ = x[TInsertValues-24]
	_ = x[TIntpSelect-25]
	_ = x[TIntpStar-26]
	_ = x[TIntpFrom-27]
	_ = x[TIntpIntoOutfile-28]
	_ = x[TIntpShowTables-29]
	_ = x[TIntpShowColumns-30]
	_ = x[TIntpShowCreateTable-31]
	_ = x[TIntpQuit-32]
	_ = x[TBeginFullStmt-33]
	_ = x[TCommitFullStmt-34]
}

const _lexItemType_name = "TIllegalTEofTSemiTCommaTCommentTDelimTNullTStringTNumberTIdentifierTDropTableFullStmtTLockTableFullStmtTUnlockTablesFullStmtTSetFullStmtTLParenTRParenTCreateTableTCreateTableDetailTCreateTableExtraDetailTColumnTypeTColumnSizeTColumnEnumValTColumnDetailsTInsertIntoTInsertValuesTIntpSelectTIntpStarTIntpFromTIntpIntoOutfileTIntpShowTablesTIntpShowColumnsTIntpShowCreateTableTIntpQuitTBeginFullStmtTCommitFullStmt"

var _lexItemType_index = [...]uint16{0, 8, 12, 17, 23, 31, 37, 42, 49, 56, 67, 85, 103, 124, 136, 143, 150, 162, 180, 203, 214, 225, 239, 253, 264, 277, 288, 297, 306, 322, 337, 353, 373, 382, 396, 411}

func (i lexItemType) String() string {
	if i >= lexItemType(len(_lexItemType_index)-1) {
		return "lexItemType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _lexItemType_name[_lexItemType_index[i]:_lexItemType_index[i+1]]
}
