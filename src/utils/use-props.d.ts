import type { Ref } from "vue";

declare module "@/utils/use-props.js" {
  type Deserializer<T> = (v: unknown) => T;
  type Serializer<T> = (v: T) => unknown;
  type PropSpec<T> = [string, Deserializer<T>?, Serializer<T>?];

  export default function useProps(
    src: Record<string, unknown>,
    mapping: Record<string, PropSpec<unknown>>
  ): [Record<string, Ref<unknown>>, () => Record<string, unknown>];
}
