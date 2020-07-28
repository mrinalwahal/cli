const { Command } = require("@oclif/command");
const yaml = require("js-yaml");
const checkForHasura = require("../util/dependencies");
const spinnerWith = require("../util/spinner");
const {
  authFileExists,
  readAuthFile,
  getCustomApiEndpoint,
} = require("../util/config");
const { validateAuth } = require("../util/login");
const chalk = require("chalk");

const fs = require("fs");
const util = require("util");
const exec = util.promisify(require("child_process").exec);
const readFile = util.promisify(fs.readFile);
const exists = util.promisify(fs.exists);

class DeployCommand extends Command {
  async run() {
    const dotNhost = "./.nhost";
    const apiUrl = getCustomApiEndpoint();
    try {
      await checkForHasura();
    } catch (err) {
      this.log(err.message);
      this.exit(1);
    }

    // check if auth file exists
    if (!(await authFileExists())) {
      this.log(
        `${chalk.red(
          "No credentials found!"
        )} Please login first with ${chalk.bold.underline("nhost login")}`
      );
      this.exit(1);
    }

    // get auth config
    const auth = readAuthFile();
    let userData;
    try {
      userData = await validateAuth(apiUrl, auth);
    } catch (err) {
      this.log(`${chalk.red("Error!")} ${err.message}`);
      this.exit(1);
    }

    if (!(await exists(`${dotNhost}`))) {
      this.log(
        `${chalk.red(
          "Error!"
        )} this directory doesn't seem to be a valid project, please run ${chalk.underline.bold(
          "nhost init"
        )} to initialize it`
      );
      this.exit(1);
    }

    const projectConfig = yaml.safeLoad(
      await readFile(`${dotNhost}/nhost.yaml`, { encoding: "utf8" })
    );
    const projectID = projectConfig.project_id;

    const project = userData.user.projects.find(
      (project) => project.id === projectID
    );

    if (!project) {
      this.log(
        `${chalk.red("Error!")} we couldn't find this project in our system`
      );
      this.exit(1);
    }

    const hasuraEndpoint = `https://${project.project_domain.hasura_domain}`;
    const adminSecret = project.hasura_gqe_admin_secret;
    try {
      let { spinner } = spinnerWith("deploying migrations");
      await exec(
        `hasura migrate apply --endpoint=${hasuraEndpoint} --admin-secret=${adminSecret}`
      );
      spinner.succeed("migrations deployed");

      ({ spinner } = spinnerWith("deploying metadata"));
      await exec(
        `hasura metadata apply --endpoint=${hasuraEndpoint} --admin-secret=${adminSecret}`
      );
      spinner.succeed("metadata deployed");
    } catch (err) {
      this.log(`\n${chalk.red("Error!")} ${err.message}`);
      this.exit(1);
    }
  }
}

DeployCommand.description = `Deploy local migrations to Nhost production
...
Deploy local migrations to Nhost production
`;

module.exports = DeployCommand;