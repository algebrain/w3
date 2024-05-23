import { describe, expect, it } from 'vitest'
import fs from "fs";
import path from "path";
import chp from "child_process";
import w3 from '../src';

const testqueryPath = path.resolve("../../utils/testquery");
const testquery = testqueryPath + "/testquery" + (process.platform === "win32" ? ".exe" : "");
console.log("SEARCH TESTQUERY IN", testquery)
if (!fs.existsSync(testquery)) throw new Error("please compile (go build) w3/utils/testquery");

const sendQuery = (query: any) => {
  return chp.spawnSync(testquery, [w3.json(query)], {encoding: "utf-8"});
}

const runQuery = (query: any) => {
  console.log("QUERY:", JSON.stringify(query, null, "  "));
  const s = sendQuery(query);
  if (s.stderr) console.log("ERROR:", s.stderr);
  console.log("RESULT:", s.stdout);
  return s.stdout;
}

describe('should', () => {
  const query = w3.search(
    w3.or(
      w3.int.is("grade", "==", 66),
      w3.int.is("age", ">=", 20),
    )
  );

  const s = runQuery(query);
  it('exported', () => {
    expect(typeof s).toBe("string");
  })
})
