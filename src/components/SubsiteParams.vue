<script setup>
import { ref } from "vue";

const props = defineProps({ params: Object, fileProps: Object });

const railSticky = ref();

function saveParams() {
  return {
    schedule_for: props.params.scheduleFor,
    data: {
      ...props.params.data,
      ...railSticky.value.saveData(),
    },
  };
}

defineExpose({
  saveParams,
});
</script>

<template>
  <SiteParamsRailSticky
    ref="railSticky"
    :file-props="fileProps"
    :params="params"
    :open="true"
  ></SiteParamsRailSticky>
</template>
