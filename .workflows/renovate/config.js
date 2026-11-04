module.exports = {
  platform: 'forgejo',
  endpoint: `${process.env.CI_FORGE_URL}/api/v1`,
  repositories: [process.env.CI_REPO],
  token: process.env.CI_NETRC_PASSWORD,
};
