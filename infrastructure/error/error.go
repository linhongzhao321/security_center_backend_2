package error

import (
	"fmt"
)

type TypeError struct {
	Expect string
	Actual string
}

func (te *TypeError) Error() string {
	return fmt.Sprintf(`type error. expected: %s, actual: %s`, te.Expect, te.Actual)
}

type OverflowError struct {
	UpperLimit string
	Actual     string
}

func (oe *OverflowError) Error() string {
	return fmt.Sprintf(`overflow error. upper limit: %s, actual: %s`, oe.UpperLimit, oe.Actual)
}

type ChanClosed struct {
	Deposit string
	Content string
}

func (chClosed *ChanClosed) Error() string {
	return fmt.Sprintf(`chanel is closed, write to chanel fail. deposit: %s, content: %T`,
		chClosed.Deposit, chClosed.Content)
}
