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

type SortDirection = "ASC" | "DESC";

interface SortQuery {
	Col: string;
	Dir: SortDirection;
}

export type Key = number | string;

export interface QueryBase {
	Params?: {[key: string]: any}; //дополнительные параметры запроса, вне логики SQL
}

export interface SelectQuery extends QueryBase {
	Limit?:  number;
	Offset?: number;
	Search:  Condition;
	Sort?:   SortQuery[];
}

export interface InsertQuery extends QueryBase {
	Insert: {
		Cols :  string[];
		Values: Value[][];
	};
}

export interface UpdateQuery extends QueryBase {
	Update: {
		Cols :  string[];
		Values: Value[][];
	}
}

export interface DeleteQuery extends QueryBase {
	Delete: Key[];
}

const createAtomaryConditionBuilder = (t: TypeName) => {
  let type_: TypeName | RangeTypeName = t;

  const is = (col: string, op: Op, val: Value | Value[]): AtomaryCondition => {
    if (op === "between") {
      if (NumericTypes.has(type_ as TypeName)) {
        type_ = "numeric";
        if (typeof val !== "number") throw new Error("[AtomaryConditionBuilder.op] number expected");
        return {
          Col:  col,
          Type: type_,
          Val:  val,
          Op:   op,
        };
      }
      
      if (!TimeTypes.has(type_ as TypeName)) {
        throw new Error("[AtomaryConditionBuilder.op] date or number type of condition expected");
      }

      if (typeof val !== "string") throw new Error("[AtomaryConditionBuilder.op] string for date expected");

      return {
        Col:  col,
        Type: type_,
        Val:  val,
        Op:   op,
      };
    }

    switch (type_) {
      case "text":
      case "textis":
      case "string":
        if (typeof val !== "string") throw new Error("[AtomaryConditionBuilder.op] string value expected");
        break;
      case "list":
        if (!Array.isArray(val)) throw new Error("[AtomaryConditionBuilder.op] array value expected");
        break;
      case "number":
      case "int":
      case "float":
        if (typeof val !== "number") throw new Error("[AtomaryConditionBuilder.op] number value expected");
        break;
      case "date":
      case "datetime":
        if (typeof val !== "string") throw new Error("[AtomaryConditionBuilder.op] string value expected");
        break;
    }
 
    return {
      Col:  col,
      Type: type_,
      Val:  val,
      Op:   op,
    };
  }

  return {
    is,
  }
}

export type AtomaryConditionBuilder = ReturnType<typeof createAtomaryConditionBuilder>;

const createW3Lib = () => {
  return {
    get number() { return createAtomaryConditionBuilder("number"); },
    get text() { return createAtomaryConditionBuilder("text"); },
    get textis() { return createAtomaryConditionBuilder("textis"); },
    get list() { return createAtomaryConditionBuilder("list"); },
    get string() { return createAtomaryConditionBuilder("string"); },
    get int() { return createAtomaryConditionBuilder("int"); },
    get float() { return createAtomaryConditionBuilder("float"); },
    get date() { return createAtomaryConditionBuilder("date"); },
    get datetime() { return createAtomaryConditionBuilder("datetime"); },

    or(...args: Condition[]): CompoundCondition {
      return { Op: "OR", Query: args }
    },
    and(...args: Condition[]): CompoundCondition {
      return { Op: "AND", Query: args }
    },
    not(...args: Condition[]): CompoundCondition {
      return { Op: "NOT", Query: args }
    },

    asc(col: string): SortQuery { return {Col: col, Dir: "ASC"} },
    desc(col: string): SortQuery { return {Col: col, Dir: "DESC"} },
    all(): SelectQuery { return { Search: this.and() } },
    search(cond: Condition, offset?: number, limit?: number, ...sort: SortQuery[]): SelectQuery {
      const r: SelectQuery = { Search: cond };
      if (offset !== undefined) r.Offset = offset;
      if (limit !== undefined) r.Limit = limit;
      if (sort.length > 0) r.Sort = sort;
      return r;
    },
    json(q: QueryBase) { return JSON.stringify(q); },
  }
};

const w3 = createW3Lib();
export default w3;