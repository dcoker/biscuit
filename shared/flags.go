package shared

import (
	"fmt"
	"strings"

	"regexp"

	"github.com/dcoker/biscuit/algorithms"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	// ProgName is the name of this program.
	ProgName = "biscuit"
)

// Command types have a Run() method.
type Command interface {
	Run() error
}

// CommaSeparatedList is a configurable flag.Value that parses a comma delimited string into a string array.
type CommaSeparatedList struct {
	V,
	allowedValues []string
	min  int
	name string
}

// Set is called by the flag parser.
func (k *CommaSeparatedList) Set(input string) error {
	valuesAllowed := make(map[string]bool)
	for _, value := range k.allowedValues {
		valuesAllowed[value] = true
	}

	k.V = strings.Split(strings.TrimSpace(input), ",")
	for i := range k.V {
		k.V[i] = strings.TrimSpace(k.V[i])
		if len(valuesAllowed) > 0 && !valuesAllowed[k.V[i]] {
			return fmt.Errorf("%s: '%s' is not a valid value", k.name, k.V[i])
		}
	}
	if len(k.V) < k.min {
		return fmt.Errorf("%syou must specify at least one", k.name)
	}
	return nil
}

// Name allows you to set a name that will be prefixed to error messages.
func (k *CommaSeparatedList) Name(name string) *CommaSeparatedList {
	k.name = name + ": "
	return k
}

// String returns the values joined with a comma.
func (k *CommaSeparatedList) String() string {
	return strings.Join(k.V, ",")
}

// RestrictTo validates that each element is a member of the provided values.
func (k *CommaSeparatedList) RestrictTo(values ...string) *CommaSeparatedList {
	k.allowedValues = values
	return k
}

// Min applies a minimum length validation to the array.
func (k *CommaSeparatedList) Min(min int) *CommaSeparatedList {
	k.min = min
	return k
}

// StringValue is a flag.Value supporting various validation behaviors.
type StringValue struct {
	v,
	name string
	trim bool
	minLength,
	maxLength *int
	regex *regexp.Regexp
}

// Set is called by the flag parser.
func (c *StringValue) Set(input string) error {
	c.v = input
	if c.trim {
		c.v = strings.TrimSpace(c.v)
	}
	if c.minLength != nil && len(c.v) < *c.minLength {
		return fmt.Errorf("%smust be at least %d characters", c.name, *c.minLength)
	}
	if c.maxLength != nil && len(c.v) > *c.maxLength {
		return fmt.Errorf("%smust be no more than %d characters", c.name, *c.maxLength)
	}
	if c.regex != nil && !c.regex.MatchString(c.v) {
		return fmt.Errorf("%smust satisfy regex %s", c.name, c.regex.String())
	}
	return nil
}

// Name allows you to set a name that will be prefixed to error messages.
func (c *StringValue) Name(name string) *StringValue {
	c.name = strings.TrimSpace(name) + ": "
	return c
}

// String returns the current flag value.
func (c *StringValue) String() string {
	return c.v
}

// MinLength validates that the value is at least of length n.
func (c *StringValue) MinLength(n int) *StringValue {
	c.minLength = &n
	return c
}

// MaxLength validates that the value is at most length n.
func (c *StringValue) MaxLength(n int) *StringValue {
	c.maxLength = &n
	return c
}

// Regex validates that the string satisfies a regular expression.
func (c *StringValue) Regex(s string) *StringValue {
	c.regex = regexp.MustCompile(s)
	return c
}

// Trimmed removes whitespace from the value before validation.
func (c *StringValue) Trimmed() *StringValue {
	c.trim = true
	return c
}

// StringFlag sets a StringValue as the target value for a Kingpin flag.
func StringFlag(s kingpin.Settings, sv *StringValue) *string {
	s.SetValue(sv)
	return &sv.v
}

// AlgorithmFlag defines a flag for the algorithm
func AlgorithmFlag(cc *kingpin.CmdClause) *string {
	return cc.Flag("algorithm", "Encryption algorithm. If the environment variable BISCUIT_ALGORITHM is "+
		"set, it will be used as the default value. Options: "+
		strings.Join(algorithms.GetAlgorithms(), ", ")).
		Short('a').
		Envar("BISCUIT_ALGORITHM").
		Default(algorithms.GetDefaultAlgorithm()).
		Enum(algorithms.GetAlgorithms()...)
}

// FilenameFlag defines a flag for the filename.
func FilenameFlag(cc *kingpin.CmdClause) *string {
	return cc.Flag("filename", "Name of file storing the secrets. If the environment variable BISCUIT_FILENAME "+
		"is set, it will be used as the default value.").
		PlaceHolder("FILE").
		Envar("BISCUIT_FILENAME").
		Short('f').
		Required().
		String()
}

// AwsRegionPriority defines a flag allowing the user to specify an ordered list of
// AWS regions to prioritize.
func AwsRegionPriorityFlag(cc *kingpin.CmdClause) *[]string {
	name := "aws-region-priority"
	fc := cc.Flag(name,
		"Comma-delimited list of AWS regions to prefer for "+
			"decryption operations. Biscuit will attempt to use the "+
			"KMS endpoints in these regions before trying the "+
			"other regions. If the environment variable AWS_REGION "+
			"is set, it will be used as the default value.").
		Short('p').
		Envar("AWS_REGION")
	val := (&CommaSeparatedList{}).Name(name)
	fc.SetValue(val)
	return &val.V
}

// SecretNameArg defines a flag for the name of the secret.
func SecretNameArg(cc *kingpin.CmdClause) *string {
	return cc.Arg("name", "Name of the secret to read.").Required().String()
}
