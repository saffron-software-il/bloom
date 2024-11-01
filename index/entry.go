package index

type EntryType int

const (
	Annotation EntryType = iota
	Attribute
	Binding
	Builtin
	Callback
	Category
	Class
	Command
	Component
	Constant
	Constructor
	Define
	Delegate
	Diagram
	Directive
	Element
	Entry
	Enum
	Environment
	Error
	Event
	Exception
	Extension
	Field
	File
	Filter
	Framework
	Function
	Global
	Guide
	Hook
	Instance
	Instruction
	Interface
	Keyword
	Library
	Literal
	Macro
	Method
	Mixin
	Modifier
	Module
	Namespace
	Notation
	Object
	Operator
	Option
	Package
	Parameter
	Plugin
	Procedure
	Property
	Protocol
	Provider
	Provisioner
	Query
	Record
	Resource
	Sample
	Section
	Service
	Setting
	Shortcut
	Statement
	Struct
	Style
	Subroutine
	Tag
	Test
	Trait
	Type
	Union
	Value
	Variable
	Word
)

type IndexEntry struct {
	Name string
	Path string
	Type EntryType
}

func NewIndexEntry(name, path string, t EntryType) IndexEntry {
	return IndexEntry{
		Name: name,
		Path: path,
		Type: t,
	}
}
