type Op = "==" | "or" | "<=" | ">=" | "<" | ">" | ">= or 0" | "!=" | "between" | "reverse in" | "in" | "not in" | "starts with" | "contains" | "ends with";
type TypeName = "text" | "textis" | "list" | "string" | "number" | "int" | "float" | "date" | "datetime";
declare const NumericTypes: ReadonlySet<TypeName>;
declare const TimeTypes: ReadonlySet<TypeName>;
declare const TextTypes: ReadonlySet<TypeName>;
type Value = number | string | boolean;
interface AtomaryCondition {
    Col: string;
    Type: string;
    Val: Value | Value[];
    Op: Op;
}
type Logics = "OR" | "AND" | "NOT";
interface CompoundCondition {
    Op: Logics;
    Query: Condition[];
}
type Condition = AtomaryCondition | CompoundCondition;
type SortDirection = "ASC" | "DESC";
interface SortQuery {
    Col: string;
    Dir: SortDirection;
}
type Key = number | string;
interface QueryBase {
    Params?: {
        [key: string]: any;
    };
}
interface SelectQuery extends QueryBase {
    Limit?: number;
    Offset?: number;
    Search: Condition;
    Sort?: SortQuery[];
}
interface InsertQuery extends QueryBase {
    Insert: {
        Cols: string[];
        Values: Value[][];
    };
}
interface UpdateQuery extends QueryBase {
    Update: {
        Cols: string[];
        Values: Value[][];
    };
}
interface DeleteQuery extends QueryBase {
    Delete: Key[];
}
declare const createAtomaryConditionBuilder: (t: TypeName) => {
    is: (col: string, op: Op, val: Value | Value[]) => void;
    readonly cond: AtomaryCondition;
};
type AtomaryConditionBuilder = ReturnType<typeof createAtomaryConditionBuilder>;
declare const w3: {
    readonly number: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly text: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly textis: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly list: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly string: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly int: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly float: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly date: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    readonly datetime: {
        is: (col: string, op: Op, val: Value | Value[]) => void;
        readonly cond: AtomaryCondition;
    };
    or(...args: Condition[]): CompoundCondition;
    and(...args: Condition[]): CompoundCondition;
    not(...args: Condition[]): CompoundCondition;
    asc(col: string): SortQuery;
    desc(col: string): SortQuery;
    all(): SelectQuery;
    search(cond: Condition, offset?: number, limit?: number, ...sort: SortQuery[]): SelectQuery;
};

export { type AtomaryCondition, type AtomaryConditionBuilder, type CompoundCondition, type Condition, type DeleteQuery, type InsertQuery, type Key, type Logics, NumericTypes, type Op, type QueryBase, type SelectQuery, TextTypes, TimeTypes, type TypeName, type UpdateQuery, type Value, w3 as default };
