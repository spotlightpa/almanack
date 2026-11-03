<script setup>
import { ref } from "vue";

const props = defineProps({ params: Object, fileProps: Object });

const railTop = ref();

function saveParams() {
  return {
    schedule_for: props.params.scheduleFor,
    data: {
      ...props.params.data,
      ...railTop.value.saveData(),
    },
  };
}

defineExpose({
  saveParams,
});
</script>

<template>
  <SiteParamsRailTop
    ref="railTop"
    :file-props="fileProps"
    :params="params"
    :open="true"
    :hideDemo="true"
  ></SiteParamsRailTop>
</template>
