<script>
import useProps from "@/utils/use-props.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      newsletterActive: ["newsletter-active"],
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
    <summary class="title is-4">Newsletter popup</summary>
    <BulmaField
      label="Newsletter"
      help="Pop up is full screen newsletter takeover"
    >
      <div>
        <label class="checkbox">
          <input v-model="newsletterActive" type="checkbox" />
          Show newsletter pop up
        </label>
      </div>
    </BulmaField>
  </details>
</template>
