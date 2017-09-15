package sqlread

import "log"

func StartIntpState(l *lexer) state {
	log.Println("a")
	l.accept(sep)
	log.Println("b")

	log.Println("c")
	if l.hasPrefixI("S") && l.hasPrefixI("SHOW") {
		l.pos += 4
		if l.accept(whitespace) < 1 || !l.hasPrefixI("TABLES") {
			l.emit(TIllegal)
			return nil
		}

		return untilSemiStateBuilder(TIntpShowTables, StartIntpState)
	}
	log.Println("d")
	if (l.hasPrefixI("Q") || l.hasPrefixI("E")) && (l.hasPrefixI("QUIT") || l.hasPrefixI("EXIT")) {
		return untilSemiStateBuilder(TIntpQuit, StartIntpState)
	}

	if l.hasPrefixI("S") && l.hasPrefixI("SELECT") {
		return selectIntpState
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

	if l.hasPrefixI("from") {
		return selectFromIntpState
	}

	c := l.next()
	if c == coma {
		l.emit(TComma)
		return selectIdentifierIntpState
	}

	return StartIntpState
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
	if c != semi {
		l.emit(TIllegal)
		return nil
	}

	l.emit(TSemi)
	return StartIntpState
}
