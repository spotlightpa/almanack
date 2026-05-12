// Type-level tests for comma-and. Verified by `tsc --noEmit`.
import commaAnd from "./comma-and.ts";

type Expect<T extends true> = T;
type Equal<X, Y> =
  (<T>() => T extends X ? 1 : 2) extends (<T>() => T extends Y ? 1 : 2) ? true : false;

// Return type must be string.
type _r1 = Expect<Equal<ReturnType<typeof commaAnd>, string>>;

// Accepts arrays of values with toString (numbers, strings, objects with toString).
const _a: string = commaAnd(["a", "b", "c"]);
const _b: string = commaAnd([1, 2, 3]);
const _c: string = commaAnd([]);
const _d: string = commaAnd(null);
const _e: string = commaAnd(undefined);

// @ts-expect-error -- bare strings are not arrays.
commaAnd("nope");

// @ts-expect-error -- numbers are not arrays.
commaAnd(42);
