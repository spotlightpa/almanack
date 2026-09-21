// Result is a Go-style [value, error] tuple for async operations.
// Exactly one of the two elements is non-null.
export type Result<T> = [T, null] | [null, Error];

// tryTo wraps a Promise so it always resolves to a Result tuple,
// avoiding try/catch at every call site.
export default function tryTo<T>(promise: Promise<T>): Promise<Result<T>> {
  return promise
    .then((data): [T, null] => [data, null])
    .catch((error: Error): [null, Error] => [null, error]);
}
