package dnsloader

const (
	// valid dnsRR_Header.Rrtype and dnsQuestion.qtype
	dnsTypeA     = 1
	dnsTypeNS    = 2
	dnsTypeMD    = 3
	dnsTypeMF    = 4
	dnsTypeCNAME = 5
	dnsTypeSOA   = 6
	dnsTypeMB    = 7
	dnsTypeMG    = 8
	dnsTypeMR    = 9
	dnsTypeNULL  = 10
	dnsTypeWKS   = 11
	dnsTypePTR   = 12
	dnsTypeHINFO = 13
	dnsTypeMINFO = 14
	dnsTypeMX    = 15
	dnsTypeTXT   = 16
	dnsTypeAAAA  = 28
	dnsTypeSRV   = 33
	dnsTypeAXFR  = 252
	dnsTypeMAILB = 253
	dnsTypeMAILA = 254
	dnsTypeALL   = 255

	// valid dnsQuestion.qclass
	dnsClassINET   = 1
	dnsClassCSNET  = 2
	dnsClassCHAOS  = 3
	dnsClassHESIOD = 4
	dnsClassANY    = 255

	// dnsMsg.rcode
	dnsRcodeSuccess        = 0
	dnsRcodeFormatError    = 1
	dnsRcodeServerFailure  = 2
	dnsRcodeNameError      = 3
	dnsRcodeNotImplemented = 4
	dnsRcodeRefused        = 5
)

var DNSType = map[string]int{
	"dnsTypeA":     1,
	"dnsTypeNS":    2,
	"dnsTypeMD":    3,
	"dnsTypeMF":    4,
	"dnsTypeCNAME": 5,
	"dnsTypeSOA":   6,
	"dnsTypeMB":    7,
	"dnsTypeMG":    8,
	"dnsTypeMR":    9,
	"dnsTypeNULL":  10,
	"dnsTypeWKS":   11,
	"dnsTypePTR":   12,
	"dnsTypeHINFO": 13,
	"dnsTypeMINFO": 14,
	"dnsTypeMX":    15,
	"dnsTypeTXT":   16,
	"dnsTypeAAAA":  28,
	"dnsTypeSRV":   33,
	"dnsTypeAXFR":  252,
	"dnsTypeMAILB": 253,
	"dnsTypeMAILA": 254,
	"dnsTypeALL":   255,
}

var QClass = map[string]int{
	"dnsClassINET":   1,
	"dnsClassCSNET":  2,
	"dnsClassCHAOS":  3,
	"dnsClassHESIOD": 4,
	"dnsClassANY":    255,
}

var DNSRcode = map[string]int{
	"dnsRcodeSuccess":        0,
	"dnsRcodeFormatError":    1,
	"dnsRcodeServerFailure":  2,
	"dnsRcodeNameError":      3,
	"dnsRcodeNotImplemented": 4,
	"dnsRcodeRefused":        5,
}
