/**
 * Custom ESLint rule: require Vue ref/reactive/computed declarations to use const.
 *
 * Catches the common mistake of `let x = ref(0)` where reassigning `x` loses
 * reactivity instead of updating the ref's `.value`.
 */

const VUE_REACTIVE_CALLS = new Set([
  "ref",
  "shallowRef",
  "reactive",
  "shallowReactive",
  "computed",
  "readonly",
  "shallowReadonly",
]);

/** Return the callee name for a CallExpression node, or null. */
function calleeName(node) {
  if (node.callee.type === "Identifier") {
    return node.callee.name;
  }
  return null;
}

export default {
  meta: {
    type: "suggestion",
    fixable: "code",
    docs: {
      description: "Require Vue ref/reactive/computed declarations to use const",
    },
    messages: {
      useConst:
        "Vue reactive value '{{name}}' should be declared with const, not let or var.",
    },
    schema: [],
  },
  create(context) {
    return {
      VariableDeclaration(node) {
        if (node.kind === "const") return;
        for (let decl of node.declarations) {
          if (
            decl.init?.type === "CallExpression" &&
            VUE_REACTIVE_CALLS.has(calleeName(decl.init))
          ) {
            const name =
              decl.id.type === "Identifier" ? decl.id.name : "(destructured)";
            context.report({
              node: decl,
              messageId: "useConst",
              data: { name },
              fix(fixer) {
                // Only safe to fix if every declarator in this statement is a
                // reactive call — otherwise changing kind affects the others.
                const allReactive = node.declarations.every(
                  (d) =>
                    d.init?.type === "CallExpression" &&
                    VUE_REACTIVE_CALLS.has(calleeName(d.init))
                );
                if (!allReactive) return null;
                return fixer.replaceText(
                  context.sourceCode.getFirstToken(node),
                  "const"
                );
              },
            });
          }
        }
      },
    };
  },
};
