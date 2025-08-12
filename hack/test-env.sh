#!/bin/bash

set -e

echo "--- Starting Colima"
colima start --kubernetes

echo "--- Installing Crossplane"
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update
helm install crossplane --namespace crossplane-system --create-namespace crossplane-stable/crossplane || echo "Already installed, skipping"
sleep 10

echo "--- Waiting for Crossplane to be ready"
kubectl wait --for=condition=Ready pods --all -n crossplane-system --timeout=600s

echo "--- Installing Crossplane CLI"
curl -sL https://raw.githubusercontent.com/crossplane/crossplane/master/install.sh | sh
sudo mv crossplane /usr/local/bin/

echo "--- Applying dummy composition and claim"
kubectl apply -f hack/fixtures

echo "--- Done"
