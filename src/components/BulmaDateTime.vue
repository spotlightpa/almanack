<script setup>
import { computed } from "vue";
import { formatDateTime } from "@/utils/time-format.js";

const props = defineProps({
  label: String,
  help: String,
  modelValue: Date,
  icon: [Array, String],
  required: Boolean,
  disabled: Boolean,
});

const emit = defineEmits(["update:modelValue"]);

const leftPad = (n) => String(n).padStart(2, "0");

const localDay = (d) => {
  if (!d) {
    return "";
  }
  let y = leftPad(d.getFullYear());
  let m = leftPad(d.getMonth() + 1);
  let dd = leftPad(d.getDate());
  return `${y}-${m}-${dd}`;
};

const day = computed(() => localDay(props.modelValue));

const time = computed(() => {
  let d = props.modelValue;
  if (!d) {
    return "";
  }
  let h = leftPad(d.getHours());
  let m = leftPad(d.getMinutes());
  let s = leftPad(d.getSeconds());

  return `${h}:${m}:${s}`;
});

function emitDay(value) {
  if (!value) {
    emit("update:modelValue", null);
    return;
  }
  let t = time.value || "00:00:00";
  emit("update:modelValue", new Date(`${value}T${t}`));
}

function emitTime(value) {
  if (!day.value && !value) {
    emit("update:modelValue", null);
    return;
  } else if (!day.value) {
    let today = localDay(new Date());
    emit("update:modelValue", new Date(`${today}T${value}`));
    return;
  }
  if (!value) {
    value = "00:00:00";
  }
  emit("update:modelValue", new Date(`${day.value}T${value}`));
}
</script>

<template>
  <BulmaField
    v-slot="{ idForLabel }"
    :label="label"
    :help="help"
    :required="required"
  >
    <div class="my-0 field has-addons">
      <p class="control is-expanded" :class="{ 'has-icons-left': !!icon }">
        <input
          :id="idForLabel"
          class="input"
          type="date"
          :value="day"
          :disabled="disabled || null"
          @input="emitDay($event.target.value)"
        />
        <span v-if="icon" class="icon is-left">
          <font-awesome-icon :icon="icon" />
        </span>
      </p>
      <p class="control is-expanded">
        <input
          class="input"
          type="time"
          :value="time"
          :disabled="disabled || null"
          step="1"
          @input="emitTime($event.target.value)"
        />
      </p>
      <p class="control">
        <span class="button is-static">
          {{ formatDateTime(modelValue) || "Unset" }}
        </span>
      </p>
    </div>
    <slot />
  </BulmaField>
</template>
