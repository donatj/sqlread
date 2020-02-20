package sqlread

import (
	"log"
	"unicode"
)

func StartState(l *lexer) state {
	l.accept(sep)

	if l.hasPrefix("--") {
		return doubleDashCommentStateBuilder(StartState)
	}

	if l.hasPrefix("/*") {
		return blockCommentStateBuilder(StartState)
	}

	if l.hasPrefix("DROP TABLE ") {
		return untilSemiStateBuilder(TDropTableFullStmt, StartState)
	}

	if l.hasPrefix("LOCK TABLES ") {
		return untilSemiStateBuilder(TLockTableFullStmt, StartState)
	}

	if l.hasPrefix("UNLOCK TABLES") {
		return untilSemiStateBuilder(TUnlockTablesFullStmt, StartState)
	}

	if l.hasPrefix("CREATE TABLE ") {
		return createTableState
	}

	if l.hasPrefix("INSERT INTO ") {
		return insertIntoTableState
	}

	if l.hasPrefix("DELIMITER ") { // TODO: Add delimiter awareness
		return untilNewlineStateBuilder(TDelim, StartState)
	}

	if l.hasPrefix("SET ") {
		return untilSemiStateBuilder(TSetFullStmt, StartState)
	}

	if l.hasPrefix("BEGIN") {
		log.Println("begin")
		return untilSemiStateBuilder(TBeginFullStmt, StartState)
	}

	if l.hasPrefix("COMMIT") {
		return untilSemiStateBuilder(TCommitFullStmt, StartState)
	}

	_, p := l.peek(1)
	if p[0] != eof {
		l.emit(TIllegal)
	}

	return nil
}

func insertIntoTableState(l *lexer) state {
	l.start = l.pos
	l.pos += 12

	l.emit(TInsertInto)
	return identifierStateBuilder(insertValuesState)
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
		l.emit(TComma)
		l.pos++
		return insertRowState
	}

	if e == semi {
		l.emit(TSemi)
		return StartState
	}

	l.emit(TIllegal)
	return nil
}

func insertRowState(l *lexer) state {
	l.start = l.pos
	for {
		c := l.next()

		if in(c, numbers) || c == dash {
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

func doubleDashCommentStateBuilder(ret state) state {
	return func(l *lexer) state {
		l.start = l.pos + 3

		for {
			b := l.next()
			if b == eof || b == lf {
				break
			}
		}

		l.emit(TComment)

		return ret
	}
}

func blockCommentStateBuilder(ret state) state {
	return func(l *lexer) state {
		l.pos += 2
		l.start = l.pos

		for {
			if l.hasPrefix("*/") {
				l.emit(TComment)
				l.pos += 2
				break
			}
			c := l.next()
			if c == eof {
				l.emit(TIllegal)
				return nil
			}
		}

		return ret
	}
}

func untilSemiStateBuilder(emit lexItemType, ret state) state {
	return func(l *lexer) state {
		l.start = l.pos

		if l.until(semi) {
			l.rewind()
			l.emit(emit)
			l.start = l.pos
			l.pos++
			l.emit(TSemi)
		} else {
			l.emit(TIllegal)
			return nil
		}

		return ret
	}
}

func untilNewlineStateBuilder(emit lexItemType, ret state) state {
	return func(l *lexer) state {
		l.start = l.pos

		if l.until('\n') {
			l.rewind()
			l.emit(emit)
			l.start = l.pos
			l.pos++
			l.emit(TSemi)
		} else {
			l.emit(TIllegal)
			return nil
		}

		return ret
	}
}

func createTableState(l *lexer) state {
	l.start = l.pos
	l.pos += 13
	l.emit(TCreateTable)

	return identifierStateBuilder(createTableParamsState)
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

	if l.hasPrefix("PRIMARY KEY ") || l.hasPrefix("KEY ") || l.hasPrefix("CONSTRAINT ") || l.hasPrefix("UNIQUE KEY") || l.hasPrefix(")") {
		return createTableDetailState
	}

	return identifierStateBuilder(createTableParamTypeState)
}

func createTableDetailState(l *lexer) state {
	l.start = l.pos
	o := 0
	for {
		c := l.next()
		if c == eof {
			l.emit(TIllegal)
			return nil
		}

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
		if c == eof {
			l.emit(TIllegal)
			return nil
		}

		if l.hasPrefix("COMMENT=") {
			l.pos += 8
			if !eatString(l) {
				l.emit(TIllegal)
				return nil
			}
		}

		if c == semi {
			l.rewind()
			l.emit(TcreateTableExtraDetail)

			l.start = l.pos
			l.next()
			l.emit(TSemi)
			return StartState
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
		if c == eof {
			l.emit(TIllegal)
			return nil
		}

		if c == coma || c == rprn {
			l.rewind()
			l.emit(TColumnDetails)
			if c == coma {
				l.next()
			}

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
}

func eatNumber(l *lexer) bool {
	if l.hasPrefix("-") {
		l.pos++
	}

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

	for {
		c := l.next()

		if c == eof {
			return false
		}

		if c == bs {
			_, p := l.peek(1)
			if p[0] == delim || p[0] == bs {
				l.next()
			}
		} else if c == delim {
			_, p := l.peek(1)
			if p[0] == delim {
				l.next()
			} else {
				break
			}

		}
	}

	return true
}

func eatIdentifier(l *lexer) bool {
	b := l.next()
	l.rewind()

	if b == bt {
		return eatDelimStr(l, bt)
	}

	log.Println("non-backtick identifiers not implemented yet at: ", l.pos)
	l.emit(TIllegal)

	// https://dev.mysql.com/doc/refman/5.7/en/identifiers.html

	return false
}

func identifierStateBuilder(ret state) state {
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
