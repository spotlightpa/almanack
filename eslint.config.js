import globals from "globals";
import pluginJs from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
import tseslint from "typescript-eslint";
import vueRefConst from "./eslint-rules/vue-ref-const.js";

import eslintConfigPrettier from "eslint-config-prettier";
import eslintPluginPrettierRecommended from "eslint-plugin-prettier/recommended";

const localRules = {
  plugins: { local: { rules: { "vue-ref-const": vueRefConst } } },
  rules: { "local/vue-ref-const": "error" },
};

export default [
  { files: ["**/*.{js,mjs,cjs,vue,ts}"] },
  { languageOptions: { globals: globals.browser } },
  pluginJs.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs["flat/essential"],
  {
    files: ["**/*.vue"],
    languageOptions: {
      parserOptions: {
        parser: tseslint.parser,
      },
    },
  },
  localRules,
  eslintPluginPrettierRecommended,
  eslintConfigPrettier,
  {
    rules: {
      "no-unused-vars": "off",
      "@typescript-eslint/no-unused-vars": ["error", { caughtErrors: "none" }],
      "prefer-const": "off",
    },
  },
  {
    files: ["**/*.test-d.ts"],
    rules: {
      "@typescript-eslint/no-unused-vars": "off",
    },
  },
];
