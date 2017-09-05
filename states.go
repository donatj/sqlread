package sqlread

import (
	"log"
	"unicode"
)

func startState(l *lexer) state {
	l.accept(sep)

	if l.hasPrefix("--") {
		return doubleDashCommentState(l, startState)
	}

	if l.hasPrefix("/*") {
		return blockCommentState(l, startState)
	}

	if l.hasPrefix("DROP TABLE ") {
		return dropTableState(l)
	}

	if l.hasPrefix("LOCK TABLES ") {
		return lockTableState(l)
	}

	if l.hasPrefix("UNLOCK TABLES") {
		return unlockTableState(l)
	}

	if l.hasPrefix("CREATE TABLE ") {
		return createTableState(l)
	}

	if l.hasPrefix("INSERT INTO ") {
		return insertIntoTableState(l)
	}

	_, p := l.peak(100)
	log.Println("peak ahead", string(p))

	return nil
}

func insertIntoTableState(l *lexer) state {
	l.start = l.pos
	l.pos += 12

	l.emit(TInsertInto)
	return identifierStateAction(insertValuesState)
}

func insertValuesState(l *lexer) state {
	l.start = l.pos
	l.accept(whitespace)
	if l.hasPrefix("VALUES ") {
		l.pos += 7
		l.emit(TInsertValues)
	} else {
		l.emit(TIllegal)
		return nil
	}

	l.accept(whitespace)

	return insertRowsState
}

func insertRowsState(l *lexer) state {
	l.start = l.pos

	e := l.next()
	if e == lprn {
		return insertRowState
	}
	// _, pp := l.peak(5)
	// log.Println(string(e), string(pp))
	if e == coma && l.hasPrefix("(") {
		l.pos++
		return insertRowState
	}

	if e == semi {
		l.emit(TSemi)
		return startState
	}

	l.emit(TIllegal)
	return nil
}

func insertRowState(l *lexer) state {
	l.start = l.pos
	for {
		c := l.next()

		if in(c, numbers) {
			l.rewind()
			if eatNumber(l) {
				l.emit(TNumber)
				break
			}
		} else if c == sq || c == dq {
			l.rewind()
			if eatString(l) {
				l.emit(TString)
				break
			}
		} else if c == letN && l.hasPrefix("ULL") {
			l.pos += 3
			l.emit(TNull)
			break
		}

		l.emit(TIllegal)
		return nil
	}

	d := l.next()

	if d == rprn {
		// l.rewind()
		// l.emit(TInsertRow)
		return insertRowsState
	}

	if d == coma {
		return insertRowState
	}

	l.emit(TIllegal)
	return nil
}

func doubleDashCommentState(l *lexer, ret state) state {
	l.start = l.pos + 3
	s := ""
	for {
		b := l.next()
		if b == eof || b == lf {
			break
		}
		s += string(b)
	}

	l.emit(TComment)

	return ret
}

func blockCommentState(l *lexer, ret state) state {
	l.pos += 2
	l.start = l.pos

	for {
		if l.hasPrefix("*/") {
			l.emit(TComment)
			l.pos += 2
			break
		}
		l.next()
	}

	return ret
}

func dropTableState(l *lexer) state {
	l.start = l.pos

	if l.until(semi) {
		l.rewind()
		l.emit(TDropTableFullStmt)
		l.start = l.pos
		l.pos++
		l.emit(TSemi)
	} else {
		l.emit(TIllegal)
		return nil
	}

	return startState
}

func lockTableState(l *lexer) state {
	l.start = l.pos

	if l.until(semi) {
		l.rewind()
		l.emit(TLockTableFullStmt)
		l.start = l.pos
		l.pos++
		l.emit(TSemi)
	} else {
		l.emit(TIllegal)
		return nil
	}

	return startState
}

func unlockTableState(l *lexer) state {
	l.start = l.pos

	if l.until(semi) {
		l.rewind()
		l.emit(TUnlockTablesFullStmt)
		l.start = l.pos
		l.pos++
		l.emit(TSemi)
	} else {
		l.emit(TIllegal)
		return nil
	}

	return startState
}

func createTableState(l *lexer) state {
	l.start = l.pos
	l.pos += 13
	l.emit(TCreateTable)

	return identifierStateAction(createTableParamsState)
}

func createTableParamsState(l *lexer) state {
	l.start = l.pos
	l.accept(whitespace)

	if l.next() == lprn {
		l.emit(TLParen)
		return createTableParamState
	}

	l.emit(TIllegal)
	return nil
}

func createTableParamState(l *lexer) state {
	l.accept(whitespace)

	if l.hasPrefix("PRIMARY KEY ") || l.hasPrefix("KEY ") || l.hasPrefix("CONSTRAINT ") {
		return createTableDetailState
	}

	return identifierStateAction(createTableParamTypeState)
}

func createTableDetailState(l *lexer) state {
	l.start = l.pos
	o := 0

	for {
		c := l.next()

		if c == lprn {
			o++
		} else if c == rprn {
			o--
		}

		if o < 0 {
			l.rewind()
			l.emit(TCreateTableDetail)

			l.start = l.pos

			l.next()
			l.emit(TRParen)
			return createTableExtra
		}
	}
}

func createTableExtra(l *lexer) state {
	l.start = l.pos
	for {
		c := l.next()

		if l.hasPrefix("COMMENT=") {
			l.pos += 8
			if !eatString(l) {
				l.emit(TIllegal)
				return nil
			}
		}

		if c == semi {
			return startState
		}
	}
}

func createTableParamTypeState(l *lexer) state {
	l.accept(whitespace)
	l.start = l.pos

	for {
		c := l.next()
		r := rune(c)

		if !unicode.IsLetter(r) {
			if l.start == l.pos-1 {
				l.emit(TIllegal)
				return nil
			}

			l.rewind()
			li := l.emit(TColumnType)

			if c == lprn {
				if li.Val == "enum" {
					return createTableParamTypeEnumValuesState
				}

				return createTableParamTypeSizeState
			}

			return createTableParamDetailsState
		}
	}
}

func createTableParamTypeEnumValuesState(l *lexer) state {
	if l.accept([]byte{lprn}) != 1 {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos
	for {
		if eatString(l) {
			l.emit(TColumnEnumVal)
		}

		c := l.next()
		if c == rprn {
			return createTableParamDetailsState
		}

		if c != coma {
			l.emit(TIllegal)
			return nil
		}

		l.start = l.pos
	}
}

func createTableParamTypeSizeState(l *lexer) state {
	if l.accept([]byte{lprn}) != 1 {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos

	for {
		c := l.next()

		if c == rprn {
			if l.start == l.pos-1 {
				l.emit(TIllegal)
				return nil
			}

			l.rewind()
			l.emit(TColumnSize)
			l.next()

			return createTableParamDetailsState
		}

		if c != coma {
			r := rune(c)

			if !unicode.IsNumber(r) {
				l.emit(TIllegal)
				return nil
			}
		}
	}
}

func createTableParamDetailsState(l *lexer) state {
	l.start = l.pos
	for {
		c := l.next()

		if c == coma || c == rprn {
			l.rewind()
			l.emit(TColumnDetails)
			l.next()
			return createTableParamState
		}

		if l.hasPrefix("COMMENT ") {
			l.pos += 7
			l.accept(whitespace)
			if !eatString(l) {
				l.emit(TIllegal)
				return nil
			}
		}

		if l.hasPrefix("DEFAULT ") {
			l.pos += 7
			l.accept(whitespace)
			eatString(l)
		}
	}
	return nil
}

func eatNumber(l *lexer) bool {
	n1 := l.accept(numbers)
	n2 := 0

	p := l.next()
	if p == dot {
		n2 = l.accept(numbers)
		if n2 == 0 {
			return false
		}
	} else {
		l.rewind()
	}

	return n1+n2 > 0
}

func eatString(l *lexer) bool {
	delim := l.next()
	l.rewind()

	if delim != sq && delim != dq {
		return false
	}

	return eatDelimStr(l, delim)
}

func eatDelimStr(l *lexer, delim byte) bool {
	first := l.next()
	if first != delim {
		l.rewind()
		return false
	}

	last := eof

	for {
		c := l.next()

		if c == eof {
			return false
		}

		if c == delim && last != bs {
			_, p := l.peak(1)
			if p[0] != delim {
				break
			} else {
				l.next()
			}
		}

		last = c
	}

	return true
}

func eatIdentifier(l *lexer) bool {
	b := l.next()
	l.rewind()

	if b == bt {
		return eatDelimStr(l, bt)
	}

	log.Fatal("non-backtick identifiers not implemented yet")

	return false
}

func identifierStateAction(ret state) state {
	return func(l *lexer) state {
		l.start = l.pos

		if eatIdentifier(l) {
			l.start++
			l.pos--
			l.emit(TIdentifier)
			l.pos++
		} else {
			l.emit(TIllegal)
		}

		return ret
	}
}
