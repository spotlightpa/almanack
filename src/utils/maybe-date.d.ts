declare module "@/utils/maybe-date.js" {
  export default function maybeDate(
    obj: unknown,
    pathStr?: string
  ): Date | null;
}
