module.exports = {
  extends: [
    "@commitlint/config-conventional"
  ],
  parserPreset: {
    parserOpts: { headerPattern: /^([^\(\):]*)(?:\((.*)\))?!?: (.*)$/ }
  },
  rules: {
    "type-enum": [
      0,
      "always"
    ],
    "subject-case": [
      2,
      "always",
      [
        "sentence-case",
        "lower-case"
      ]
    ],
    "body-leading-blank": [
      2,
      "always"
    ],
    "footer-leading-blank": [
      2,
      "always"
    ]
  }
};
