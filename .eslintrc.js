module.exports = {
  root: true,
  env: {
    node: true
  },
  extends: ["plugin:vue/essential", "@vue/prettier"],
  rules: {
    "no-console": "warn",
    "no-debugger": "warn"
  },
  parserOptions: {
    parser: "babel-eslint"
  }
};
