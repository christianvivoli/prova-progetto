package app

import (
	"fmt"
	"os"

	"github.com/go-stack/stack"
	log "github.com/inconshreveable/log15"
)

// LogErr si occupa di eseguire il log nel canale più appropriato in base al codice d'errore passato, inoltre è possibile definire un ulteriore skip che si andrà a sommare a quello attuale '1' nel caso in cui questa funzione viene richiamata all'interno di un wrapper.
func LogErr(logger log.Logger, err error, addSkip ...int) {

	skip := 1

	if len(addSkip) > 0 {
		skip += addSkip[0]
	}

	caller := stack.Caller(skip)

	msg := ErrorMessage(err)
	errEntries := ErrorLogEntries(err)

	ctxs := append(errEntries,
		"log_file", fmt.Sprint(caller),
		"log_fn", fmt.Sprintf("%+n", caller))

	switch ErrorCode(err) {
	case
		EINTERNAL,
		EINTERNAL_INVALID,
		EUNKNOWN,
		ECONFLICT,
		ENOTINJECTED:
		logger.Error(msg, ctxs...)

	case
		ENOTIMPLEMENTED,
		ECANCELED,
		EFORBIDDEN:
		logger.Warn(msg, ctxs...)

	default:
		logger.Info(msg, ctxs...)
	}
}

func MultiHandlerLogger() log.Handler {
	return log.MultiHandler(
		log.StreamHandler(os.Stdout, log.TerminalFormat()),
	)
}
