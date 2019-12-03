module.exports = {
  env: {
    browser: true,
    es6: true,
    jest: true
  },
  extends: ["airbnb-typescript", "prettier", "prettier/@typescript-eslint", "prettier/react"],
  globals: {
    Atomics: "readonly",
    SharedArrayBuffer: "readonly"
  },
  parserOptions: {
    ecmaFeatures: {
      jsx: true
    },
    ecmaVersion: 2018,
    sourceType: "module"
  },
  plugins: ["react"],
  rules: {
    "import/prefer-default-export": "off",
    "max-len": "off",
    "no-console": "off",
    "@typescript-eslint/no-unused-vars": "off",
    "react/destructuring-assignment": "off",
    "no-plusplus": "off",
    "default-case": "off",
    "react/prop-types": "off",
    "react/jsx-props-no-spreading" : "off",
    "import/order" : "off"
  }
}
