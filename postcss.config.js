const purgecss = require("@fullhuman/postcss-purgecss");

let plugins = [require("autoprefixer")];
if (process.env.NODE_ENV === "production") {
  plugins.push(
    purgecss({
      content: ["src/**/*.html", "src/**/*.vue"],
    })
  );
}

module.exports = {
  plugins,
};
