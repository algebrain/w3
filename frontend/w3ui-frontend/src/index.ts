export type Op = "==" | "or" | "<=" | ">=" | "<" | ">" | ">= or 0" | "!=" | 
  "between" | "reverse in" | "in" | "not in" | "starts with" | "contains" | 
  "ends with";

export type Value = number | string | boolean;

export interface AtomaryCondition {
  Col:  string;
  Type: string;
  Val:  Value | Value[];
  Op:   Op;
}

export type Logics = "OR" | "AND" | "NOT";

export interface CompoundCondition {
	Op:    Logics;
	Query: Condition[];
}

export type Condition = AtomaryCondition | CompoundCondition;

interface SortQuery {
	Col: string;
	Dir: string;
}

export type Key = number | string;

export interface Query {
	Limit?: number;
	Offset?: number;
	Search?: Condition;
	Sort?:   SortQuery[];
	Insert?: {
		Cols :  string[];
		Values: Value[][];
	};
	Update?: {
		Cols :  string[];
		Values: Value[][];
	}
	Delete?: Key[];
	Params?: {[key: string]: any}; //дополнительные параметры запроса, вне логики SQL
}


