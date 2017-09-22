package sqlread

func StartIntpState(l *lexer) state {
	l.accept(sep)

	if l.hasPrefixI("S") && l.hasPrefixI("SHOW") {
		l.pos += 4
		if l.accept(whitespace) < 1 {
			l.emit(TIllegal)
			return nil
		}

		if l.hasPrefixI("TABLES") {
			return untilSemiStateBuilder(TIntpShowTables, nil)
		}
		if l.hasPrefixI("COLUMNS") {
			return showColumnsIntpState
		}

		l.emit(TIllegal)
		return nil
	}

	if (l.hasPrefixI("Q") || l.hasPrefixI("E")) && (l.hasPrefixI("QUIT") || l.hasPrefixI("EXIT")) {
		return untilSemiStateBuilder(TIntpQuit, nil)
	}

	if l.hasPrefixI("S") && l.hasPrefixI("SELECT") {
		return selectIntpState
	}

	return nil
}

func showColumnsIntpState(l *lexer) state {
	l.pos += 7
	l.emit(TIntpShowColumns)

	if l.accept(whitespace) < 1 {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos

	if !l.hasPrefixI("FROM") {
		l.emit(TIllegal)
		return nil
	}
	l.pos += 4
	l.emit(TIntpFrom)

	if l.accept(whitespace) < 1 {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos

	if eatIdentifier(l) {
		l.emit(TIdentifier)
	} else {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos

	c := l.next()
	if c != semi {
		l.emit(TIllegal)
		return nil
	}

	return nil
}

func selectIntpState(l *lexer) state {
	l.start = l.pos
	l.pos += 6

	l.emit(TIntpSelect)

	if l.accept(whitespace) < 1 {
		l.emit(TIllegal)
		return nil
	}

	return selectIdentifierIntpState
}

func selectIdentifierIntpState(l *lexer) state {
	l.accept(whitespace)
	l.start = l.pos

	if l.hasPrefix("*") {
		l.pos += 1
		l.emit(TIntpStar)
	} else if eatIdentifier(l) {
		l.emit(TIdentifier)
	} else {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos

	l.accept(whitespace)

	if l.hasPrefixI("FROM") {
		return selectFromIntpState
	}

	c := l.next()
	if c == coma {
		l.emit(TComma)
		return selectIdentifierIntpState
	}

	// return StartIntpState
	l.emit(TIllegal)
	return nil
}

func selectFromIntpState(l *lexer) state {
	l.start = l.pos
	l.pos += 4
	l.emit(TIntpFrom)

	if l.accept(whitespace) < 1 {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos
	if !eatIdentifier(l) {
		l.emit(TIllegal)
		return nil
	}

	l.emit(TIdentifier)

	l.accept(whitespace)

	l.start = l.pos

	c := l.next()
	if c == byte('i') || c == byte('I') {
		if l.hasPrefixI("NTO") {
			l.pos += 3
			if l.accept(whitespace) < 1 {
				l.emit(TIllegal)
				return nil
			}
			if l.hasPrefixI("O") && l.hasPrefixI("OUTFILE") {
				l.pos += 7
				return selectFromIntoOutfileIntpState
			}
		}
	}

	if c != semi {
		l.emit(TIllegal)
		return nil
	}

	l.emit(TSemi)
	return nil
}

func selectFromIntoOutfileIntpState(l *lexer) state {
	l.emit(TIntpIntoOutfile)
	l.start = l.pos

	if l.accept(whitespace) < 1 {
		l.emit(TIllegal)
		return nil
	}

	l.start = l.pos
	if !eatString(l) {
		l.emit(TIllegal)
		return nil
	}
	l.emit(TString)

	l.accept(whitespace)

	c := l.next()
	if c != semi {
		l.emit(TIllegal)
		return nil
	}

	l.emit(TSemi)
	return nil
}
