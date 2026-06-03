#!/usr/bin/env python3
"""
deploy/tests/test_deploy_contract.py

Contract tests for all deploy-related YAML and Dockerfile artefacts:
  - Dockerfile
  - docker-compose.stage.yml / docker-compose.prod.yml
  - .github/actions/deploy/action.yml
  - .github/actions/notify/action.yml
  - .github/workflows/ci-cd.yml
  - .gitlab-ci.yml
  - .gitverse/workflows/ci-cd.yml

Run:  python3 deploy/tests/test_deploy_contract.py
Exit: 0 = all passed, non-zero = at least one failure.

No third-party test libraries. Uses unittest + yaml + re from stdlib.
"""

import os
import re
import sys
import unittest
import yaml

REPO = os.path.normpath(os.path.join(os.path.dirname(__file__), "..", ".."))


def repo(*parts: str) -> str:
    return os.path.join(REPO, *parts)


def load_yaml(rel_path: str) -> dict:
    with open(repo(rel_path)) as f:
        return yaml.safe_load(f)


def read_file(rel_path: str) -> str:
    with open(repo(rel_path)) as f:
        return f.read()


# ─────────────────────────────────────────────────────────────────────────────
# Dockerfile
# ─────────────────────────────────────────────────────────────────────────────
class DockerfileContract(unittest.TestCase):

    def setUp(self):
        self.src = read_file("Dockerfile")

    # Prevents: changing the exposed port without updating docker-compose files,
    # Nginx configs, healthcheck URLs, and CI env vars that all reference 8080.
    def test_expose_8080(self):
        self.assertIn("EXPOSE 8080", self.src,
                      "EXPOSE port changed; update docker-compose, CI HEALTHCHECK_URL, and nginx config")

    # Prevents: changing the default CMD port from 8080 (app would start on
    # a different port than compose/nginx expects, causing silent traffic loss).
    def test_cmd_serve_port_8080(self):
        self.assertRegex(self.src, r'CMD\s+\[.*"serve".*"8080"',
                         'CMD port changed; must match EXPOSE and docker-compose port mapping')

    # Prevents: changing ENTRYPOINT to something other than the binary
    # (would silently run the wrong executable in production).
    def test_entrypoint_is_binary(self):
        self.assertIn('/usr/local/bin/n8go-docs', self.src,
                      "ENTRYPOINT binary path changed; update deploy docs and runbooks")

    # Prevents: removing the VOLUME declarations (docs/themes would be baked into
    # the image and updates would require a full image rebuild).
    def test_volumes_declared(self):
        self.assertIn('VOLUME', self.src,
                      "VOLUME declarations removed; live content updates require mounted volumes")
        self.assertIn('/docs', self.src)
        self.assertIn('/site', self.src)

    # Prevents: removing OCI image labels (breaks image provenance tracking in
    # registries and compliance audits).
    def test_oci_labels_present(self):
        self.assertIn('org.opencontainers.image.title', self.src,
                      "OCI labels removed; image provenance tracking breaks")

    # Prevents: switching from distroless to a base image with a shell
    # (increases attack surface; distroless is a security requirement).
    def test_base_image_is_distroless(self):
        self.assertIn('distroless', self.src,
                      "Base image changed from distroless; security review required")


# ─────────────────────────────────────────────────────────────────────────────
# docker-compose files
# ─────────────────────────────────────────────────────────────────────────────
class DockerComposeStageContract(unittest.TestCase):

    def setUp(self):
        self.cfg = load_yaml("docker-compose.stage.yml")
        self.app = self.cfg["services"]["app"]

    # Prevents: renaming the compose service from "app" — provider scripts
    # hard-code SERVICE=app and would stop managing the right container.
    def test_service_named_app(self):
        self.assertIn("app", self.cfg["services"],
                      "Service renamed from 'app'; update SERVICE= in all provider scripts")

    # Prevents: removing the port mapping (app becomes unreachable from host/nginx).
    def test_port_8080_mapped(self):
        ports = self.app.get("ports", [])
        self.assertTrue(any("8080" in str(p) for p in ports),
                        "Port 8080 mapping removed from stage compose file")

    # Prevents: removing the IMAGE env-var substitution — deploy script exports
    # IMAGE before calling compose up; without it the old image tag is used.
    def test_image_uses_env_var(self):
        image = self.app.get("image", "")
        self.assertIn("${IMAGE", image,
                      "image: must reference ${IMAGE} so the deploy script can inject the tag")

    # Prevents: removing the healthcheck (docker-compose won't report container
    # state correctly; compose-level rolling updates rely on it).
    def test_healthcheck_configured(self):
        self.assertIn("healthcheck", self.app,
                      "healthcheck removed from stage compose file")

    # Prevents: changing restart policy to "no" (service won't recover from crashes).
    def test_restart_policy(self):
        restart = self.app.get("restart", "")
        self.assertIn(restart, ("unless-stopped", "always", "on-failure"),
                      f"restart policy '{restart}' will not recover from crashes")

    # Prevents: removing the docs/themes volume mounts (content would be baked
    # into the image and could not be updated without a full redeploy).
    def test_volumes_mounted(self):
        volumes = self.app.get("volumes", [])
        vol_str = str(volumes)
        self.assertIn("docs", vol_str,
                      "docs volume removed from stage compose file")
        self.assertIn("themes", vol_str,
                      "themes volume removed from stage compose file")


class DockerComposeProdContract(unittest.TestCase):

    def setUp(self):
        self.cfg = load_yaml("docker-compose.prod.yml")
        self.app = self.cfg["services"]["app"]

    def test_service_named_app(self):
        self.assertIn("app", self.cfg["services"],
                      "Service renamed from 'app'; update SERVICE= in all provider scripts")

    # Prevents: binding prod port to 0.0.0.0 (app would be directly reachable
    # from outside; traffic should go through nginx only).
    def test_prod_port_bound_to_localhost(self):
        ports = self.app.get("ports", [])
        for p in ports:
            if "8080" in str(p):
                self.assertIn("127.0.0.1", str(p),
                              "Prod port 8080 must bind to 127.0.0.1 only — nginx proxies from outside")
                return
        self.fail("Port 8080 not found in prod compose file")

    def test_image_uses_env_var(self):
        image = self.app.get("image", "")
        self.assertIn("${IMAGE", image,
                      "image: must reference ${IMAGE}")

    def test_healthcheck_configured(self):
        self.assertIn("healthcheck", self.app)

    # Prevents: changing prod restart to unless-stopped (prod must always restart
    # even after docker daemon restarts, e.g. after host reboot).
    def test_restart_always(self):
        self.assertEqual(self.app.get("restart"), "always",
                         "Prod restart policy must be 'always' to survive host reboots")

    # Prevents: removing the memory limit (OOM killer could take down unrelated
    # services on the host; prod containers must have resource limits).
    def test_memory_limit_set(self):
        deploy = self.app.get("deploy", {})
        resources = deploy.get("resources", {})
        limits = resources.get("limits", {})
        self.assertIn("memory", limits,
                      "Memory limit removed from prod compose file; add it to prevent OOM on host")


# ─────────────────────────────────────────────────────────────────────────────
# GitHub Actions: deploy composite action
# ─────────────────────────────────────────────────────────────────────────────
class DeployActionContract(unittest.TestCase):

    def setUp(self):
        self.cfg = load_yaml(".github/actions/deploy/action.yml")
        self.inputs = self.cfg.get("inputs", {})

    # Prevents: removing or renaming the 'provider' input — all workflow callers
    # pass provider: docker-ssh explicitly; rename breaks every pipeline silently.
    def test_provider_input_exists(self):
        self.assertIn("provider", self.inputs,
                      "Input 'provider' removed from deploy action; update all workflow callers")

    def test_provider_input_required(self):
        self.assertTrue(self.inputs["provider"].get("required"),
                        "Input 'provider' must be required=true")

    # Prevents: removing the 'environment' input (used by notify step and audit logs).
    def test_environment_input_exists(self):
        self.assertIn("environment", self.inputs,
                      "Input 'environment' removed from deploy action")

    # Prevents: removing the 'image' input (deploy action would not know what to deploy).
    def test_image_input_required(self):
        self.assertIn("image", self.inputs,
                      "Input 'image' removed from deploy action")
        self.assertTrue(self.inputs["image"].get("required"),
                        "Input 'image' must be required=true")

    # Prevents: removing SSH credential inputs (docker-ssh provider would fall back
    # to passwordless SSH and fail silently or use wrong key).
    def test_ssh_inputs_exist(self):
        for inp in ("ssh_host", "ssh_user", "ssh_port", "ssh_key"):
            self.assertIn(inp, self.inputs,
                          f"SSH input '{inp}' removed from deploy action")

    # Prevents: removing registry credential inputs (docker login would fail).
    def test_registry_inputs_exist(self):
        for inp in ("registry_server", "registry_user", "registry_password"):
            self.assertIn(inp, self.inputs,
                          f"Registry input '{inp}' removed from deploy action")
        self.assertTrue(self.inputs["registry_password"].get("required"),
                        "registry_password must be required=true")

    # Prevents: changing the default SSH port from 22 (most servers use 22;
    # changing default would break all deployments that don't override ssh_port).
    def test_ssh_port_default_is_22(self):
        default = str(self.inputs.get("ssh_port", {}).get("default", ""))
        self.assertEqual(default, "22",
                         f"ssh_port default changed to '{default}'; was 22")

    # Prevents: changing the default service name from 'app' (provider scripts
    # and docker-compose files all reference SERVICE=app).
    def test_compose_service_default_is_app(self):
        default = self.inputs.get("compose_service", {}).get("default", "")
        self.assertEqual(default, "app",
                         f"compose_service default changed to '{default}'; update provider scripts")

    # Prevents: changing the default compose dir from /opt/app (deploy scripts
    # cd to this directory; changing it silently runs compose from wrong path).
    def test_compose_project_dir_default(self):
        default = self.inputs.get("compose_project_dir", {}).get("default", "")
        self.assertEqual(default, "/opt/app",
                         f"compose_project_dir default changed to '{default}'")

    # Prevents: removing healthcheck inputs (provider script would use its own
    # defaults and the action's documented defaults would be ignored).
    def test_healthcheck_inputs_have_defaults(self):
        hc_url = self.inputs.get("healthcheck_url", {}).get("default", "")
        self.assertIn("8080", hc_url,
                      f"healthcheck_url default '{hc_url}' does not reference port 8080")

        retries = str(self.inputs.get("healthcheck_retries", {}).get("default", ""))
        self.assertTrue(retries.isdigit() and int(retries) > 0,
                        f"healthcheck_retries default must be a positive integer, got '{retries}'")

    # Prevents: removing the ssh_key install step (provider would use wrong key).
    def test_install_ssh_key_step_present(self):
        steps = self.cfg.get("runs", {}).get("steps", [])
        names = [s.get("name", "") for s in steps]
        self.assertTrue(any("ssh" in n.lower() or "key" in n.lower() for n in names),
                        "SSH key install step missing from deploy action")

    # Prevents: removing the credentials cleanup step (SSH key lingers on runner disk).
    def test_cleanup_step_present(self):
        steps = self.cfg.get("runs", {}).get("steps", [])
        cleanup = [s for s in steps if s.get("if") == "always()"]
        self.assertTrue(len(cleanup) > 0,
                        "No always() cleanup step in deploy action; SSH key and SA key may leak")

    # Prevents: renaming supported providers without updating the action router.
    # All workflows pass provider: docker-ssh; if the step condition changes they stop deploying.
    def test_docker_ssh_provider_step_exists(self):
        steps = self.cfg.get("runs", {}).get("steps", [])
        docker_ssh_steps = [s for s in steps if "docker-ssh" in str(s.get("if", ""))]
        self.assertTrue(len(docker_ssh_steps) > 0,
                        "No step conditioned on provider==docker-ssh; all docker-ssh deploys would be skipped")


# ─────────────────────────────────────────────────────────────────────────────
# GitHub Actions: notify composite action
# ─────────────────────────────────────────────────────────────────────────────
class NotifyActionContract(unittest.TestCase):

    def setUp(self):
        self.cfg = load_yaml(".github/actions/notify/action.yml")
        self.inputs = self.cfg.get("inputs", {})

    # Prevents: removing 'status' input (notification would have no status field).
    def test_status_input_required(self):
        self.assertIn("status", self.inputs)
        self.assertTrue(self.inputs["status"].get("required"))

    # Prevents: removing Telegram credential inputs (curl call would fail with 401).
    def test_telegram_inputs_required(self):
        for inp in ("telegram_token", "telegram_chat_id"):
            self.assertIn(inp, self.inputs,
                          f"Input '{inp}' removed from notify action")
            self.assertTrue(self.inputs[inp].get("required"),
                            f"Input '{inp}' must be required=true")

    def test_environment_and_image_inputs(self):
        for inp in ("environment", "image", "provider"):
            self.assertIn(inp, self.inputs,
                          f"Input '{inp}' removed from notify action")


# ─────────────────────────────────────────────────────────────────────────────
# GitHub Actions: CI/CD workflow
# ─────────────────────────────────────────────────────────────────────────────
class GitHubWorkflowContract(unittest.TestCase):

    def setUp(self):
        self.cfg = load_yaml(".github/workflows/ci-cd.yml")
        self.jobs = self.cfg.get("jobs", {})
        self.env = self.cfg.get("env", {})

    # Prevents: removing the concurrency block (concurrent deploys to same env
    # would corrupt state — rolling update interrupted by another rolling update).
    def test_concurrency_block_present(self):
        self.assertIn("concurrency", self.cfg,
                      "concurrency block removed; concurrent deploys will corrupt deployment state")

    def test_concurrency_cancel_in_progress_is_false(self):
        # Prevents: setting cancel-in-progress: true on the deploy concurrency group.
        # If a deploy is cancelled mid-way the service is left in a broken state.
        val = self.cfg.get("concurrency", {}).get("cancel-in-progress")
        self.assertFalse(val,
                         "cancel-in-progress must be false for deploy; cancelling mid-deploy leaves service broken")

    # Prevents: renaming the HEALTHCHECK_URL env var (provider scripts reference it).
    def test_healthcheck_url_env_var(self):
        self.assertIn("HEALTHCHECK_URL", self.env,
                      "HEALTHCHECK_URL env var removed from workflow; provider scripts will use wrong URL")
        url = self.env["HEALTHCHECK_URL"]
        self.assertIn("8080", url,
                      f"HEALTHCHECK_URL '{url}' does not reference port 8080")

    # Prevents: removing required jobs from the pipeline.
    def test_required_jobs_present(self):
        for job in ("lint", "test", "build", "docker", "deploy-stage", "deploy-prod"):
            self.assertIn(job, self.jobs,
                          f"Job '{job}' removed from CI/CD workflow")

    # Prevents: removing the test→build dependency (build could succeed even if
    # tests fail, shipping broken code).
    def test_build_needs_test(self):
        needs = self.jobs.get("build", {}).get("needs", [])
        if isinstance(needs, str):
            needs = [needs]
        self.assertIn("test", needs,
                      "build job no longer requires test; broken code could be built and shipped")

    # Prevents: removing the build→docker dependency (docker job could use a
    # stale binary artifact from a previous run).
    def test_docker_needs_build(self):
        needs = self.jobs.get("docker", {}).get("needs", [])
        if isinstance(needs, str):
            needs = [needs]
        self.assertIn("build", needs,
                      "docker job no longer requires build; could use stale binary artifact")

    # Prevents: accidentally removing the pull_request guard on the docker push job.
    # Without it, every PR would push an image to the registry.
    def test_docker_skipped_on_pr(self):
        condition = self.jobs.get("docker", {}).get("if", "")
        self.assertIn("pull_request", condition,
                      "docker job 'if' condition does not exclude pull_request; PRs would push images")

    # Prevents: changing the stage deploy trigger from main branch to something else.
    def test_deploy_stage_triggers_on_main(self):
        condition = self.jobs.get("deploy-stage", {}).get("if", "")
        self.assertIn("main", condition,
                      "deploy-stage 'if' condition no longer checks for main branch")

    # Prevents: changing the prod deploy trigger from semver tags to something else.
    def test_deploy_prod_triggers_on_version_tag(self):
        condition = self.jobs.get("deploy-prod", {}).get("if", "")
        self.assertTrue(re.search(r"tags/v|ref_name.*v|startsWith.*v", condition),
                        f"deploy-prod condition '{condition}' no longer checks for semver tag")

    # Prevents: removing the packages: write permission (GHCR push would fail with 403).
    def test_packages_write_permission(self):
        perms = self.cfg.get("permissions", {})
        self.assertEqual(perms.get("packages"), "write",
                         "packages: write permission removed; docker push to GHCR will fail")

    # Prevents: renaming the IMAGE_NAME env var (docker job computes tags from it;
    # renaming would break the image tag computation).
    def test_image_name_env_var(self):
        self.assertIn("IMAGE_NAME", self.env,
                      "IMAGE_NAME env var removed from workflow")


# ─────────────────────────────────────────────────────────────────────────────
# GitLab CI
# ─────────────────────────────────────────────────────────────────────────────
class GitLabCIContract(unittest.TestCase):

    def setUp(self):
        self.cfg = load_yaml(".gitlab-ci.yml")

    # Prevents: reordering stages — dependency between jobs is implied by stage order;
    # reordering could run deploy before tests pass.
    def test_stage_order(self):
        stages = self.cfg.get("stages", [])
        required = ["lint", "test", "build", "docker", "deploy", "notify"]
        for stage in required:
            self.assertIn(stage, stages,
                          f"Stage '{stage}' removed from GitLab CI")

        # deploy must come after docker
        idx_docker = stages.index("docker")
        idx_deploy = stages.index("deploy")
        self.assertGreater(idx_deploy, idx_docker,
                           "deploy stage must come after docker stage")

    # Prevents: changing the default image from golang:* (test/lint/build jobs
    # would run without Go installed and fail with confusing errors).
    def test_default_image_is_golang(self):
        default_img = self.cfg.get("default", {}).get("image", "")
        self.assertIn("golang", default_img,
                      f"default image '{default_img}' is not a Go image; test/build jobs will fail")

    # Prevents: renaming the deploy job for stage — GitLab environment URL and
    # notify job `needs:` reference it by name.
    def test_deploy_stage_job_exists(self):
        self.assertIn("deploy:stage", self.cfg,
                      "deploy:stage job renamed; update notify job needs: and environment URLs")

    def test_deploy_prod_job_exists(self):
        self.assertIn("deploy:prod", self.cfg,
                      "deploy:prod job renamed; update environment configs and runbooks")

    # Prevents: removing the manual approval gate from prod (automated push to prod).
    def test_deploy_prod_is_manual(self):
        prod_job = self.cfg.get("deploy:prod", {})
        self.assertEqual(prod_job.get("when"), "manual",
                         "deploy:prod 'when: manual' removed; prod would deploy automatically without approval")

    # Prevents: removing the HEALTHCHECK_URL variable (provider script would use
    # its own default which may differ from the documented endpoint).
    def test_healthcheck_url_variable(self):
        variables = self.cfg.get("variables", {})
        self.assertIn("HEALTHCHECK_URL", variables,
                      "HEALTHCHECK_URL variable removed from .gitlab-ci.yml")

    # Prevents: removing the deploy.env artifact passing (docker job writes
    # PRIMARY_IMAGE to deploy.env; deploy jobs source it to get the image tag).
    def test_docker_job_exports_primary_image(self):
        docker_job = self.cfg.get("docker:build-push", {})
        artifacts = docker_job.get("artifacts", {})
        reports = artifacts.get("reports", {})
        dotenv = reports.get("dotenv", "")
        self.assertIn("deploy.env", dotenv,
                      "docker:build-push no longer exports deploy.env; deploy jobs will not get PRIMARY_IMAGE")


# ─────────────────────────────────────────────────────────────────────────────
# GitVerse CI
# ─────────────────────────────────────────────────────────────────────────────
class GitVerseCIContract(unittest.TestCase):

    def setUp(self):
        self.cfg = load_yaml(".gitverse/workflows/ci-cd.yml")
        self.jobs = self.cfg.get("jobs", {})
        self.env = self.cfg.get("env", {})

    def test_required_jobs_present(self):
        for job in ("lint", "test", "build", "docker", "deploy-stage", "deploy-prod"):
            self.assertIn(job, self.jobs,
                          f"Job '{job}' removed from GitVerse CI/CD workflow")

    # Prevents: removing the gitverse-specific REGISTRY env var
    # (would fall back to an undefined variable and push to the wrong registry).
    def test_registry_env_var(self):
        self.assertIn("REGISTRY", self.env,
                      "REGISTRY env var removed from GitVerse workflow")
        registry = self.env["REGISTRY"]
        self.assertIn("gitverse", registry,
                      f"REGISTRY '{registry}' does not reference gitverse registry")

    # Prevents: removing the manual approval on prod (automated prod deploy).
    def test_deploy_prod_uses_environment(self):
        prod = self.jobs.get("deploy-prod", {})
        self.assertIn("environment", prod,
                      "deploy-prod has no environment block; manual approval gate is missing")

    def test_docker_skipped_on_pr(self):
        condition = self.jobs.get("docker", {}).get("if", "")
        self.assertIn("pull_request", condition,
                      "GitVerse docker job does not exclude pull_request")


# ─────────────────────────────────────────────────────────────────────────────
# Cross-platform consistency
# ─────────────────────────────────────────────────────────────────────────────
class CrossPlatformConsistency(unittest.TestCase):
    """
    Verifies that key constants are consistent across all CI platforms.
    Drift between platforms causes prod to behave differently than stage on one
    platform vs another — the most dangerous class of CI/CD bug.
    """

    def _hc_url(self, cfg_path: str, variables_key: str = "env") -> str:
        cfg = load_yaml(cfg_path)
        return str(cfg.get(variables_key, {}).get("HEALTHCHECK_URL", ""))

    # Prevents: HEALTHCHECK_URL drifting between platforms so one platform
    # always reports healthy while another reports failed.
    def test_healthcheck_url_port_consistent(self):
        github_url = self._hc_url(".github/workflows/ci-cd.yml", "env")
        gitlab_url = self._hc_url(".gitlab-ci.yml", "variables")
        gitverse_url = self._hc_url(".gitverse/workflows/ci-cd.yml", "env")

        for name, url in (
            ("GitHub", github_url),
            ("GitLab", gitlab_url),
            ("GitVerse", gitverse_url),
        ):
            self.assertIn("8080", url,
                          f"{name} HEALTHCHECK_URL '{url}' does not reference port 8080")

    # Prevents: binary name drifting across platforms (would produce artefacts
    # with different names that Dockerfile COPY would not find).
    def test_binary_name_consistent(self):
        github_env = load_yaml(".github/workflows/ci-cd.yml").get("env", {})
        gitlab_vars = load_yaml(".gitlab-ci.yml").get("variables", {})
        gitverse_env = load_yaml(".gitverse/workflows/ci-cd.yml").get("env", {})

        github_name = github_env.get("BINARY_NAME", "")
        gitlab_name = gitlab_vars.get("BINARY_NAME", "")
        gitverse_name = gitverse_env.get("BINARY_NAME", "")

        self.assertEqual(github_name, gitlab_name,
                         f"BINARY_NAME mismatch: GitHub='{github_name}', GitLab='{gitlab_name}'")
        self.assertEqual(github_name, gitverse_name,
                         f"BINARY_NAME mismatch: GitHub='{github_name}', GitVerse='{gitverse_name}'")

    # Prevents: Go version drifting between platforms (would produce binaries
    # compiled with different versions — one platform could ship a regression).
    def test_golang_version_consistent_major(self):
        github_file = load_yaml(".github/workflows/ci-cd.yml")
        gitlab_file = load_yaml(".gitlab-ci.yml")

        github_img = gitlab_file.get("default", {}).get("image", "")
        # Extract version number from "golang:1.25-alpine" → "1.25"
        m_gl = re.search(r"golang:(\d+\.\d+)", github_img)
        if m_gl:
            gitlab_version = m_gl.group(1)
        else:
            # No image in gitlab = skip comparison
            return

        gitverse_env = load_yaml(".gitverse/workflows/ci-cd.yml").get("env", {})
        gitverse_version = gitverse_env.get("GO_VERSION", "")
        if gitverse_version:
            self.assertEqual(gitlab_version, str(gitverse_version),
                             f"Go version mismatch: GitLab='{gitlab_version}', GitVerse='{gitverse_version}'")


if __name__ == "__main__":
    loader = unittest.TestLoader()
    suite = unittest.TestSuite()

    for cls in [
        DockerfileContract,
        DockerComposeStageContract,
        DockerComposeProdContract,
        DeployActionContract,
        NotifyActionContract,
        GitHubWorkflowContract,
        GitLabCIContract,
        GitVerseCIContract,
        CrossPlatformConsistency,
    ]:
        suite.addTests(loader.loadTestsFromTestCase(cls))

    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)
    sys.exit(0 if result.wasSuccessful() else 1)
