package builtins

import "github.com/CWBudde/go-dws-lsp/internal/analysis"

// GetBuiltinSignature returns the signature for a built-in function if it exists
// Returns nil if the function is not a known built-in
func GetBuiltinSignature(functionName string) *analysis.FunctionSignature {
	// Check if this is a built-in function
	if sig, exists := builtinSignatures[functionName]; exists {
		return &sig
	}
	return nil
}

// builtinSignatures contains predefined signatures for DWScript built-in functions
var builtinSignatures = map[string]analysis.FunctionSignature{
	// String functions
	"PrintLn": {
		Name: "PrintLn",
		Parameters: []analysis.ParameterInfo{
			{Name: "text", Type: "String", IsOptional: false},
		},
		ReturnType:    "",
		Documentation: "Prints a line of text to the console",
	},

	"Print": {
		Name: "Print",
		Parameters: []analysis.ParameterInfo{
			{Name: "text", Type: "String", IsOptional: false},
		},
		ReturnType:    "",
		Documentation: "Prints text to the console without a newline",
	},

	"IntToStr": {
		Name: "IntToStr",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Integer", IsOptional: false},
		},
		ReturnType:    "String",
		Documentation: "Converts an integer value to a string",
	},

	"StrToInt": {
		Name: "StrToInt",
		Parameters: []analysis.ParameterInfo{
			{Name: "text", Type: "String", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Converts a string to an integer value",
	},

	"FloatToStr": {
		Name: "FloatToStr",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Float", IsOptional: false},
		},
		ReturnType:    "String",
		Documentation: "Converts a float value to a string",
	},

	"StrToFloat": {
		Name: "StrToFloat",
		Parameters: []analysis.ParameterInfo{
			{Name: "text", Type: "String", IsOptional: false},
		},
		ReturnType:    "Float",
		Documentation: "Converts a string to a float value",
	},

	"Length": {
		Name: "Length",
		Parameters: []analysis.ParameterInfo{
			{Name: "str", Type: "String", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Returns the length of a string",
	},

	"Copy": {
		Name: "Copy",
		Parameters: []analysis.ParameterInfo{
			{Name: "str", Type: "String", IsOptional: false},
			{Name: "index", Type: "Integer", IsOptional: false},
			{Name: "count", Type: "Integer", IsOptional: false},
		},
		ReturnType:    "String",
		Documentation: "Returns a substring starting at index with the specified length",
	},

	"Pos": {
		Name: "Pos",
		Parameters: []analysis.ParameterInfo{
			{Name: "subStr", Type: "String", IsOptional: false},
			{Name: "str", Type: "String", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Returns the position of a substring within a string",
	},

	"UpperCase": {
		Name: "UpperCase",
		Parameters: []analysis.ParameterInfo{
			{Name: "str", Type: "String", IsOptional: false},
		},
		ReturnType:    "String",
		Documentation: "Converts a string to uppercase",
	},

	"LowerCase": {
		Name: "LowerCase",
		Parameters: []analysis.ParameterInfo{
			{Name: "str", Type: "String", IsOptional: false},
		},
		ReturnType:    "String",
		Documentation: "Converts a string to lowercase",
	},

	"Trim": {
		Name: "Trim",
		Parameters: []analysis.ParameterInfo{
			{Name: "str", Type: "String", IsOptional: false},
		},
		ReturnType:    "String",
		Documentation: "Removes leading and trailing whitespace from a string",
	},

	// Array functions
	"SetLength": {
		Name: "SetLength",
		Parameters: []analysis.ParameterInfo{
			{Name: "arr", Type: "Array", IsOptional: false},
			{Name: "length", Type: "Integer", IsOptional: false},
		},
		ReturnType:    "",
		Documentation: "Sets the length of a dynamic array",
	},

	"High": {
		Name: "High",
		Parameters: []analysis.ParameterInfo{
			{Name: "arr", Type: "Array", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Returns the highest valid index of an array",
	},

	"Low": {
		Name: "Low",
		Parameters: []analysis.ParameterInfo{
			{Name: "arr", Type: "Array", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Returns the lowest valid index of an array",
	},

	// Math functions
	"Abs": {
		Name: "Abs",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Float",
		Documentation: "Returns the absolute value of a number",
	},

	"Sqrt": {
		Name: "Sqrt",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Float",
		Documentation: "Returns the square root of a number",
	},

	"Round": {
		Name: "Round",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Rounds a float value to the nearest integer",
	},

	"Trunc": {
		Name: "Trunc",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Truncates a float value to an integer",
	},

	"Floor": {
		Name: "Floor",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Returns the largest integer less than or equal to the value",
	},

	"Ceil": {
		Name: "Ceil",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Returns the smallest integer greater than or equal to the value",
	},

	"Sin": {
		Name: "Sin",
		Parameters: []analysis.ParameterInfo{
			{Name: "angle", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Float",
		Documentation: "Returns the sine of an angle in radians",
	},

	"Cos": {
		Name: "Cos",
		Parameters: []analysis.ParameterInfo{
			{Name: "angle", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Float",
		Documentation: "Returns the cosine of an angle in radians",
	},

	"Tan": {
		Name: "Tan",
		Parameters: []analysis.ParameterInfo{
			{Name: "angle", Type: "Float", IsOptional: false},
		},
		ReturnType:    "Float",
		Documentation: "Returns the tangent of an angle in radians",
	},

	// Type checking/conversion
	"VarType": {
		Name: "VarType",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Variant", IsOptional: false},
		},
		ReturnType:    "Integer",
		Documentation: "Returns the type code of a variant value",
	},

	"VarIsNull": {
		Name: "VarIsNull",
		Parameters: []analysis.ParameterInfo{
			{Name: "value", Type: "Variant", IsOptional: false},
		},
		ReturnType:    "Boolean",
		Documentation: "Returns true if a variant value is null",
	},

	// Date/Time functions
	"Now": {
		Name:          "Now",
		Parameters:    []analysis.ParameterInfo{},
		ReturnType:    "DateTime",
		Documentation: "Returns the current date and time",
	},

	"Date": {
		Name:          "Date",
		Parameters:    []analysis.ParameterInfo{},
		ReturnType:    "DateTime",
		Documentation: "Returns the current date",
	},

	"Time": {
		Name:          "Time",
		Parameters:    []analysis.ParameterInfo{},
		ReturnType:    "DateTime",
		Documentation: "Returns the current time",
	},
}

// IsBuiltinFunction checks if a function name is a built-in function
func IsBuiltinFunction(functionName string) bool {
	_, exists := builtinSignatures[functionName]
	return exists
}
