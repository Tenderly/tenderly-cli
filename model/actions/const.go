package actions

import (
	"regexp"

	"github.com/robfig/cron/v3"
)

var (
	RuntimeV1          = "v1"
	TriggerTypes       = []string{"periodic", "webhook", "block", "transaction", "alert"}
	PeriodicType       = "periodic"
	WebhookType        = "webhook"
	BlockType          = "block"
	TransactionType    = "transaction"
	AlertType          = "alert"
	Invocations        = []string{"any", "direct", "internal"}
	InvocationAny      = "any"
	InvocationDirect   = "direct"
	InvocationInternal = "internal"

	Intervals      = []string{"5m", "10m", "15m", "30m", "1h", "3h", "6h", "12h", "1d"}
	IntervalToCron = map[string]string{
		"5m":  "*/5 * * * *",
		"10m": "*/10 * * * *",
		"15m": "*/15 * * * *",
		"30m": "*/30 * * * *",
		"1h":  "0 * * * *",
		"3h":  "0 */3 * * *",
		"6h":  "0 */6 * * *",
		"12h": "0 */12 * * *",
		"1d":  "0 0 * * *",
	}

	CronParser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

	AddressRegex = "^0x[0-9a-f]{40}$"
	AddressRe    = regexp.MustCompile(AddressRegex)
	SigRegex     = "^0x[0-9a-f]{8}$"
	SigRe        = regexp.MustCompile(SigRegex)

	MsgTriggerTypeNotSupported             = "trigger type '%s' not supported, supported types %s"
	MsgTriggerTypeMismatch                 = "trigger type '%s' different from configured trigger"
	MsgNetworkNotSupported                 = "network '%s' is not supported"
	MsgStatusNotSupported                  = "status '%s' is not supported, supported statuses %s"
	MsgAddressDoesNotMatchRegex            = "address '%s' does not match regex %s"
	MsgSignatureDoesNotMatchRegex          = "signature '%s' does not match regex '%s'"
	MsgDefaultToAuthenticated              = "authenticated not set, defaulting to true"
	MsgIntervalOrCronRequired              = "one of 'cron' or 'interval' is required"
	MsgIntervalAndCronForbidden            = "both 'cron' and 'interval' is forbidden"
	MsgIntervalNotSupported                = "interval '%s' not supported, supported intervals %s"
	MsgCronNotSupported                    = "cron '%s' is not supported, got error %s"
	MsgBlocksNegative                      = "blocks must be greater than 0, found %d"
	MsgDefaultToAnyInvocation              = "invocation not set for contract, defaulting to any"
	MsgInvocationNotSupported              = "invocation '%s' not supported, supported invocations %s"
	MsgAccountOrContractRequired           = "one of 'account' or 'contract' is required"
	MsgAccountAndContractForbidden         = "both 'account' and 'contract' is forbidden"
	MsgContractRequired                    = "'contract' is required"
	MsgSignatureOrNameRequired             = "one of 'signature' or 'name' is required"
	MsgSignatureAndNameForbidden           = "both 'signature' and 'name' is forbidden"
	MsgSignatureAndParameterForbidden      = "'parameter' can not be used with 'signature'"
	MsgIdOrNameRequired                    = "one of 'id' or 'name' is required"
	MsgIdAndNameForbidden                  = "both 'id' and 'name' is forbidden"
	MsgIdAndParameterForbidden             = "'parameter' can not be used with 'id'"
	MsgKeyOrFieldRequired                  = "one of 'key' or 'field' is required"
	MsgKeyAndFieldForbidden                = "both 'key' and 'field' is forbidden"
	MsgKeyAndValueOrPreviousValueForbidden = "'value' or 'previousValue' can not be used with 'key'"
	MsgValueAndPreviousValueForbidden      = "both 'value' and 'previousValue' is forbidden"
	MsgTxStatusNotSupported                = "transaction status '%s' not supported, supported %s"
	MsgStatusRequired                      = "'status' must have at least one element"
	MsgFiltersRequired                     = "'filters' must have at least one element"
	MsgStartsWithEmpty                     = "'startsWith' must have at least one element"
	MsgStartsWithInvalid                   = "'startsWith' element must be hex encoded and start with 0x"
	MsgHexValueEmpty                       = "expected non-empty hex value"
	MsgHexValueInvalid                     = "hex value must start with 0x, got %s"
)
