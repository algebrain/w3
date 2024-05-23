'use strict';

Object.defineProperty(exports, '__esModule', { value: true });

const NumericTypes = /* @__PURE__ */ new Set([
  "number",
  "int",
  "float"
]);
const TimeTypes = /* @__PURE__ */ new Set([
  "date",
  "datetime"
]);
const TextTypes = /* @__PURE__ */ new Set([
  "text",
  "textis",
  "string"
]);
const createAtomaryConditionBuilder = (t) => {
  let type_ = t;
  let cond_;
  const is = (col, op, val) => {
    if (op === "between") {
      if (NumericTypes.has(type_)) {
        type_ = "numeric";
        if (typeof val !== "number")
          throw new Error("[AtomaryConditionBuilder.op] number expected");
        cond_ = {
          Col: col,
          Type: type_,
          Val: val,
          Op: op
        };
        return;
      }
      if (!TimeTypes.has(type_)) {
        throw new Error("[AtomaryConditionBuilder.op] date or number type of condition expected");
      }
      if (typeof val !== "string")
        throw new Error("[AtomaryConditionBuilder.op] string for date expected");
      cond_ = {
        Col: col,
        Type: type_,
        Val: val,
        Op: op
      };
      return;
    }
    switch (type_) {
      case "text":
      case "textis":
      case "string":
        if (typeof val !== "string")
          throw new Error("[AtomaryConditionBuilder.op] string value expected");
        break;
      case "list":
        if (!Array.isArray(val))
          throw new Error("[AtomaryConditionBuilder.op] array value expected");
        break;
      case "number":
      case "int":
      case "float":
        if (typeof val !== "number")
          throw new Error("[AtomaryConditionBuilder.op] number value expected");
        break;
      case "date":
      case "datetime":
        if (typeof val !== "string")
          throw new Error("[AtomaryConditionBuilder.op] string value expected");
        break;
    }
    cond_ = {
      Col: col,
      Type: type_,
      Val: val,
      Op: op
    };
  };
  return {
    is,
    get cond() {
      return cond_;
    }
  };
};
const createW3Lib = () => {
  return {
    get number() {
      return createAtomaryConditionBuilder("number");
    },
    get text() {
      return createAtomaryConditionBuilder("text");
    },
    get textis() {
      return createAtomaryConditionBuilder("textis");
    },
    get list() {
      return createAtomaryConditionBuilder("list");
    },
    get string() {
      return createAtomaryConditionBuilder("string");
    },
    get int() {
      return createAtomaryConditionBuilder("int");
    },
    get float() {
      return createAtomaryConditionBuilder("float");
    },
    get date() {
      return createAtomaryConditionBuilder("date");
    },
    get datetime() {
      return createAtomaryConditionBuilder("datetime");
    },
    or(...args) {
      return { Op: "OR", Query: args };
    },
    and(...args) {
      return { Op: "AND", Query: args };
    },
    not(...args) {
      return { Op: "NOT", Query: args };
    },
    asc(col) {
      return { Col: col, Dir: "ASC" };
    },
    desc(col) {
      return { Col: col, Dir: "DESC" };
    },
    all() {
      return { Search: this.and() };
    },
    search(cond, offset, limit, ...sort) {
      const r = { Search: cond };
      if (offset !== void 0)
        r.Offset = offset;
      if (limit !== void 0)
        r.Limit = limit;
      if (sort.length > 0)
        r.Sort = sort;
      return r;
    }
  };
};
const w3 = createW3Lib();

exports.NumericTypes = NumericTypes;
exports.TextTypes = TextTypes;
exports.TimeTypes = TimeTypes;
exports.default = w3;
