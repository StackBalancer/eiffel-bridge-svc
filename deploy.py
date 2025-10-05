#!/usr/bin/env python3
import subprocess, sys, getpass, os, platform, shutil

def run(cmd, background=False, output=False):
    print(f"\n>>> {cmd}")
    if background:
        # Start process in background (no blocking, logs hidden)
        return subprocess.Popen(cmd, shell=True,
                                stdout=subprocess.DEVNULL,
                                stderr=subprocess.DEVNULL)
    elif output:
        return subprocess.check_output(cmd, shell=True).decode().strip()
    else:
        subprocess.check_call(cmd, shell=True)

def ask(prompt, default=None, secret=False):
    if secret:
        val = getpass.getpass(f"{prompt} (default: {default}): ")
    else:
        val = input(f"{prompt} (default: {default}): ")
    return val.strip() or default

def check_and_install_minikube():
    if shutil.which("minikube"):
        print("minikube found")
        return
    install = ask("minikube not found. Install now? (y/n)", "y")
    if install.lower().startswith("y"):
        sysname = platform.system().lower()
        url = f"https://storage.googleapis.com/minikube/releases/latest/minikube-{sysname}-amd64"
        run(f"curl -Lo minikube {url}")
        run("chmod +x minikube && sudo mv minikube /usr/local/bin/")
    else:
        sys.exit("minikube is required, exiting.")

def check_and_install_helm():
    if shutil.which("helm"):
        print("helm found")
        return
    install = ask("helm not found. Install now? (y/n)", "y")
    if install.lower().startswith("y"):
        sysname = platform.system().lower()
        url = f"https://get.helm.sh/helm-v3.16.2-{sysname}-amd64.tar.gz"
        run(f"curl -Lo helm.tar.gz {url}")
        run("tar -xzf helm.tar.gz")
        run("sudo mv ./*/helm /usr/local/bin/")
        run("rm -rf helm.tar.gz ./*/helm")
    else:
        sys.exit("helm is required, exiting.")

def main():
    print("""
=============================================
🚀 EiffelBridge Local Test Environment Setup
=============================================

Requirements:
  • Docker or Podman installed
  • Python 3.8+
  • curl, tar, sudo available

This script will:
  1. Optionally install minikube and helm
  2. Start minikube cluster
  3. Deploy RabbitMQ + EiffelBridge svc via Helm
  4. Port-forward EiffelBridge svc and RabbitMQ UI automatically
  5. Print ready-to-use test commands
=============================================
""")

    # --- Check tools ---
    check_and_install_minikube()
    check_and_install_helm()

    # --- Cluster setup ---
    driver = ask("Minikube driver (docker/podman)", "docker")
    cpus = ask("CPUs for cluster", "2")
    memory = ask("Memory (MB)", "4096")
    ctr_runtime = ""
    network = ""

    if driver == "podman":
        ctr_runtime = "--container-runtime=crio"
        network = "--network=podman"
        print("Configuring Minikube for Podman rootless mode...")
        run("minikube config set rootless true")

    run(f"minikube start --driver={driver} {network} {ctr_runtime} --cpus={cpus} --memory={memory}")

    # --- EiffelBridge image build + push ---
    image_tag = f"eiffel-bridge"
    run(f"rm -rf {image_tag}.tar")

    if driver == "podman":
        run(f"podman build -t {image_tag} .")
        run(f"podman image save {image_tag} -o {image_tag}.tar")
    else:
        run(f"docker build -t {image_tag} .")
        run(f"docker image save {image_tag} -o {image_tag}.tar")

    # Load image into Minikube (for both drivers)
    print("Loading image into Minikube...")
    run(f"minikube image load {image_tag}.tar")

    # --- EiffelBridge deploy ---
    print("Deploying EiffelBridge...")
    run(f"helm upgrade --install eiffel-bridge ./charts/eiffel-bridge")

    print("\nWaiting for pods to become ready...")
    run("kubectl wait --for=condition=available --timeout=60s deployment/eiffel-bridge")
    run("kubectl wait --for=condition=available --timeout=60s deployment/rabbitmq")

    # --- RabbitMQ Management GUI ---
    print("\nSetting up RabbitMQ Management UI port-forward on http://localhost:15672 ...")
    run("kubectl port-forward svc/rabbitmq 15672:15672", background=True)
    print("Port-forward started in background, UI is ready.")

    # --- EiffelBridge port-forward ---
    print("\nSetting up EiffelBridge port-forward on http://localhost:8080 ...")
    run("kubectl port-forward svc/eiffel-bridge 8080:8080", background=True)
    print("EiffelBridge service is now accessible locally at http://localhost:8080/webhook")

    print("\nServices are available locally:")
    print("  EiffelBridge API → http://localhost:8080/webhook")
    print("  RabbitMQ UI      → http://localhost:15672\n")
    print("\nDeployment complete!\n")

    # --- Test instructions ---
    print("You can now test a webhook locally using:")
    print("  curl -X POST http://localhost:8080/webhook \\")
    print("       -H 'Content-Type: application/json' \\")
    print("       -H 'X-Gitlab-Event: Push Hook' \\")
    print("       --data-binary @test-webhooks/push.json\n")
    print("Other sample payloads are available in test-webhooks/ folder.")

    rabbitmqPod = run("kubectl get pods -l app=rabbitmq -o jsonpath='{.items[0].metadata.name}'", output=True)
    print("\nConsume events from RabbitMQ queue:")
    print(f"  kubectl exec -it {rabbitmqPod} -- bash \\")
    print(f"     rabbitmqadmin -u guest -p guest get queue=eiffel.events count=<num_of_events>\n")

if __name__ == "__main__":
    try:
        main()
    except subprocess.CalledProcessError as e:
        print(f"Failed: {e}")
        sys.exit(1)
