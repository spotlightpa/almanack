<script setup>
import useProps from "@/utils/use-props.js";

const props = defineProps({
  params: Object,
  fileProps: Object,
});

const [data, saveData] = useProps(props.params.data, {
  newsletterActive: ["newsletter-active"],
});

const { newsletterActive } = data;
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
