<script>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      takeoverActive: ["takeover-active"],
      takeoverHed: ["takeover-hed"],
      takeoverDek: ["takeover-dek"],
      takeoverImage: ["takeover-image"],
      takeoverLink: ["takeover-link", toAbs, toRel],
    });
    return {
      ...data,
      saveData,
    };
  },
};
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Takeover popup</summary>

    <BulmaField label="Takeover" help="Show screen obscuring takeover">
      <div>
        <label class="checkbox">
          <input v-model="takeoverActive" type="checkbox" />
          Show takeover promo
        </label>
      </div>
    </BulmaField>
    <template v-if="takeoverActive">
      <BulmaFieldInput v-model="takeoverHed" label="Hed for takeover" />
      <BulmaTextarea v-model="takeoverDek" label="Dek for takeover" />
      <BulmaFieldInput
        v-model="takeoverLink"
        label="Takeover link"
        type="url"
      />
      <BulmaField
        label="Takeover image"
        help="Must be an Almanack Photo ID"
        v-slot="{ idForLabel }"
      >
        <div class="is-flex">
          <input :id="idForLabel" v-model="takeoverImage" class="input" />
          <BulmaPaste @paste="takeoverImage = $event" />
        </div>
      </BulmaField>
    </template>
  </details>
</template>
