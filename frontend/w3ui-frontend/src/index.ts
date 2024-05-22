export type Op = "==" | "or" | "<=" | ">=" | "<" | ">" | ">= or 0" | "!=" | 
  "between" | "reverse in" | "in" | "not in" | "starts with" | "contains" | 
  "ends with";

export type TypeName = "text" | "textis" | "list" | "string" | "number" | "int" | "float" | "date" | "datetime";

type RangeTypeName = "date" | "datetime" | "numeric";

export const NumericTypes: ReadonlySet<TypeName> = new Set([
  "number", "int", "float"
]);

export const TimeTypes: ReadonlySet<TypeName> = new Set([
  "date", "datetime"
]);

export const TextTypes: ReadonlySet<TypeName> = new Set([
  "text", "textis", "string"
]);

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


class AtomaryConditionBuilder {
  private type_: TypeName | RangeTypeName;
  private cond_?: AtomaryCondition;

  constructor(t: TypeName) {
    this.type_ = t;
  };

  op(op: Op, col: string, val: Value | Value[]) {
    if (op === "between") {
      if (NumericTypes.has(this.type_ as TypeName)) {
        this.type_ = "numeric";
        if (typeof val !== "number") throw new Error("[AtomaryQueryBuilder.op] number expected");
        this.cond_ = {
          Col:  col,
          Type: this.type_,
          Val:  val,
          Op:   op,
        };
        return;
      }
      
      if (!TimeTypes.has(this.type_ as TypeName)) {
        throw new Error("[AtomaryQueryBuilder.op] date or number type of condition expected");
      }

      if (typeof val !== "string") throw new Error("[AtomaryQueryBuilder.op] string for date expected");

      this.cond_ = {
        Col:  col,
        Type: this.type_,
        Val:  val,
        Op:   op,
      };

      return;
    }

    switch (this.type_) {
      case "text":
      case "textis":
      case "string":
        if (typeof val !== "string") throw new Error("[AtomaryQueryBuilder.op] string value expected");
        break;
      case "list":
        if (!Array.isArray(val)) throw new Error("[AtomaryQueryBuilder.op] array value expected");
        break;
      case "number":
      case "int":
      case "float":
        if (typeof val !== "number") throw new Error("[AtomaryQueryBuilder.op] number value expected");
        break;
      case "date":
      case "datetime":
        if (typeof val !== "string") throw new Error("[AtomaryQueryBuilder.op] string value expected");
        break;
    }
 
    this.cond_ = {
      Col:  col,
      Type: this.type_,
      Val:  val,
      Op:   op,
    };
  }

  get cond() { return this.cond_; }
}