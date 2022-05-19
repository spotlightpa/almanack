<script>
export default {
  name: "BulmaFieldInput",
  props: {
    label: String,
    labelClass: {
      type: String,
      default: "label",
    },
    value: String,
    placeholder: String,
    help: String,
    validator: Function,
    name: String,
    required: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      validationMessage: "",
    };
  },
  computed: {
    fieldProps() {
      return {
        label: this.label,
        help: this.help,
        labelClass: this.labelClass,
        required: this.required,
        validationMessage: this.validationMessage,
      };
    },
  },
  watch: {},
  methods: {},
};
</script>

<template>
  <BulmaField ref="field" v-slot="{ idForLabel }" v-bind="fieldProps">
    <div class="is-flex is-align-items-center">
      <input
        :id="idForLabel"
        :value="value"
        type="color"
        @input="$emit('input', $event.target.value)"
      />
      <span class="ml-4 is-flex-grow-0">
        <input
          :value="value"
          :required="required"
          :name="name"
          :placeholder="placeholder"
          type="text"
          class="input is-small"
          @input="$emit('input', $event.target.value)"
        />
      </span>
      <BulmaPaste
        button-class="ml-2 button is-primary is-small has-text-weight-semibold"
        @paste="$emit('input', $event)"
      />
    </div>
  </BulmaField>
</template>
