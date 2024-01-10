<script>
export default {
  props: {
    label: String,
    labelClass: {
      type: String,
      default: "label",
    },
    modelValue: Boolean,
    placeholder: String,
    help: String,
    name: String,
  },
  computed: {
    fieldProps() {
      return {
        label: this.label,
        help: this.help,
        labelClass: this.labelClass,
      };
    },
    checkboxValue: {
      get() {
        return this.modelValue;
      },
      set(event) {
        this.$emit("update:modelValue", event);
      },
    },
  },
};
</script>

<template>
  <BulmaField v-slot="{ idForLabel }" v-bind="fieldProps">
    <div>
      <label class="checkbox">
        <input :id="idForLabel" v-model="checkboxValue" type="checkbox" />
        <slot />
      </label>
    </div>
  </BulmaField>
</template>
