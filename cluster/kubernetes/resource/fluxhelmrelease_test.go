package resource

import (
	"testing"

	"github.com/weaveworks/flux/resource"
)

func TestParseImageOnlyFormat(t *testing.T) {
	expectedImage := "bitnami/mariadb:10.1.30-r1"
	doc := `---
apiVersion: helm.integrations.flux.weave.works/v1alpha2
kind: FluxHelmRelease
metadata:
  name: mariadb
  namespace: maria
  labels:
    chart: mariadb
spec:
  chartGitPath: mariadb
  values:
    first: post
    image: ` + expectedImage + `
    persistence:
      enabled: false
`

	resources, err := ParseMultidoc([]byte(doc), "test")
	if err != nil {
		t.Fatal(err)
	}
	res, ok := resources["maria:fluxhelmrelease/mariadb"]
	if !ok {
		t.Fatalf("expected resource not found; instead got %#v", resources)
	}
	fhr, ok := res.(resource.Workload)
	if !ok {
		t.Fatalf("expected resource to be a Workload, instead got %#v", res)
	}

	containers := fhr.Containers()
	if len(containers) != 1 {
		t.Errorf("expected 1 container; got %#v", containers)
	}
	image := containers[0].Image.String()
	if image != expectedImage {
		t.Errorf("expected container image %q, got %q", expectedImage, image)
	}
}

func TestParseImageTagFormat(t *testing.T) {
	expectedImageName := "bitnami/mariadb"
	expectedImageTag := "10.1.30-r1"
	expectedImage := expectedImageName + ":" + expectedImageTag

	doc := `---
apiVersion: helm.integrations.flux.weave.works/v1alpha2
kind: FluxHelmRelease
metadata:
  name: mariadb
  namespace: maria
  labels:
    chart: mariadb
spec:
  chartGitPath: mariadb
  values:
    first: post
    image: ` + expectedImageName + `
    tag: ` + expectedImageTag + `
    persistence:
      enabled: false
`

	resources, err := ParseMultidoc([]byte(doc), "test")
	if err != nil {
		t.Fatal(err)
	}
	res, ok := resources["maria:fluxhelmrelease/mariadb"]
	if !ok {
		t.Fatalf("expected resource not found; instead got %#v", resources)
	}
	fhr, ok := res.(resource.Workload)
	if !ok {
		t.Fatalf("expected resource to be a Workload, instead got %#v", res)
	}

	containers := fhr.Containers()
	if len(containers) != 1 {
		t.Errorf("expected 1 container; got %#v", containers)
	}
	image := containers[0].Image.String()
	if image != expectedImage {
		t.Errorf("expected container image %q, got %q", expectedImage, image)
	}
}

func TestParseNamedImageFormat(t *testing.T) {
	expectedContainer := "db"
	expectedImage := "bitnami/mariadb:10.1.30-r1"
	doc := `---
apiVersion: helm.integrations.flux.weave.works/v1alpha2
kind: FluxHelmRelease
metadata:
  name: mariadb
  namespace: maria
  labels:
    chart: mariadb
spec:
  chartGitPath: mariadb
  values:
    ` + expectedContainer + `:
      first: post
      image: ` + expectedImage + `
      persistence:
        enabled: false
`

	resources, err := ParseMultidoc([]byte(doc), "test")
	if err != nil {
		t.Fatal(err)
	}
	res, ok := resources["maria:fluxhelmrelease/mariadb"]
	if !ok {
		t.Fatalf("expected resource not found; instead got %#v", resources)
	}
	fhr, ok := res.(resource.Workload)
	if !ok {
		t.Fatalf("expected resource to be a Workload, instead got %#v", res)
	}

	containers := fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image := containers[0].Image.String()
	if image != expectedImage {
		t.Errorf("expected container image %q, got %q", expectedImage, image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}

	newImage := containers[0].Image.WithNewTag("some-other-tag")
	if err := fhr.SetContainerImage(expectedContainer, newImage); err != nil {
		t.Error(err)
	}

	containers = fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image = containers[0].Image.String()
	if image != newImage.String() {
		t.Errorf("expected container image %q, got %q", newImage.String(), image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}
}

func TestParseNamedImageTagFormat(t *testing.T) {
	expectedContainer := "db"
	expectedImageName := "bitnami/mariadb"
	expectedImageTag := "10.1.30-r1"
	expectedImage := expectedImageName + ":" + expectedImageTag

	doc := `---
apiVersion: helm.integrations.flux.weave.works/v1alpha2
kind: FluxHelmRelease
metadata:
  name: mariadb
  namespace: maria
  labels:
    chart: mariadb
spec:
  chartGitPath: mariadb
  values:
    other:
      not: "containing image"
    ` + expectedContainer + `:
      first: post
      image: ` + expectedImageName + `
      tag: ` + expectedImageTag + `
      persistence:
        enabled: false
`

	resources, err := ParseMultidoc([]byte(doc), "test")
	if err != nil {
		t.Fatal(err)
	}
	res, ok := resources["maria:fluxhelmrelease/mariadb"]
	if !ok {
		t.Fatalf("expected resource not found; instead got %#v", resources)
	}
	fhr, ok := res.(resource.Workload)
	if !ok {
		t.Fatalf("expected resource to be a Workload, instead got %#v", res)
	}

	containers := fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image := containers[0].Image.String()
	if image != expectedImage {
		t.Errorf("expected container image %q, got %q", expectedImage, image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}

	newImage := containers[0].Image.WithNewTag("some-other-tag")
	if err := fhr.SetContainerImage(expectedContainer, newImage); err != nil {
		t.Error(err)
	}

	containers = fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image = containers[0].Image.String()
	if image != newImage.String() {
		t.Errorf("expected container image %q, got %q", newImage.String(), image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}
}

func TestParseNamedImageObjectFormat(t *testing.T) {
	expectedContainer := "db"
	expectedImageName := "bitnami/mariadb"
	expectedImageTag := "10.1.30-r1"
	expectedImage := expectedImageName + ":" + expectedImageTag

	doc := `---
apiVersion: helm.integrations.flux.weave.works/v1alpha2
kind: FluxHelmRelease
metadata:
  name: mariadb
  namespace: maria
  labels:
    chart: mariadb
spec:
  chartGitPath: mariadb
  values:
    other:
      not: "containing image"
    ` + expectedContainer + `:
      first: post
      image:
        repository: ` + expectedImageName + `
        tag: ` + expectedImageTag + `
      persistence:
        enabled: false
`

	resources, err := ParseMultidoc([]byte(doc), "test")
	if err != nil {
		t.Fatal(err)
	}
	res, ok := resources["maria:fluxhelmrelease/mariadb"]
	if !ok {
		t.Fatalf("expected resource not found; instead got %#v", resources)
	}
	fhr, ok := res.(resource.Workload)
	if !ok {
		t.Fatalf("expected resource to be a Workload, instead got %#v", res)
	}

	containers := fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image := containers[0].Image.String()
	if image != expectedImage {
		t.Errorf("expected container image %q, got %q", expectedImage, image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}

	newImage := containers[0].Image.WithNewTag("some-other-tag")
	if err := fhr.SetContainerImage(expectedContainer, newImage); err != nil {
		t.Error(err)
	}

	containers = fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image = containers[0].Image.String()
	if image != newImage.String() {
		t.Errorf("expected container image %q, got %q", newImage.String(), image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}
}

func TestParseNamedImageObjectFormatFlat(t *testing.T) {
	expectedContainer := ReleaseContainerName
	expectedImageName := "bitnami/mariadb"
	expectedImageTag := "10.1.30-r1"
	expectedImage := expectedImageName + ":" + expectedImageTag

	doc := `---
apiVersion: helm.integrations.flux.weave.works/v1alpha2
kind: FluxHelmRelease
metadata:
  name: mariadb
  namespace: maria
  labels:
    chart: mariadb
spec:
  chartGitPath: mariadb
  values:
    image:
      repository: ` + expectedImageName + `
      tag: ` + expectedImageTag + `
    persistence:
      enabled: false
`

	resources, err := ParseMultidoc([]byte(doc), "test")
	if err != nil {
		t.Fatal(err)
	}
	res, ok := resources["maria:fluxhelmrelease/mariadb"]
	if !ok {
		t.Fatalf("expected resource not found; instead got %#v", resources)
	}
	fhr, ok := res.(resource.Workload)
	if !ok {
		t.Fatalf("expected resource to be a Workload, instead got %#v", res)
	}

	containers := fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image := containers[0].Image.String()
	if image != expectedImage {
		t.Errorf("expected container image %q, got %q", expectedImage, image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}

	newImage := containers[0].Image.WithNewTag("some-other-tag")
	if err := fhr.SetContainerImage(expectedContainer, newImage); err != nil {
		t.Error(err)
	}

	containers = fhr.Containers()
	if len(containers) != 1 {
		t.Fatalf("expected 1 container; got %#v", containers)
	}
	image = containers[0].Image.String()
	if image != newImage.String() {
		t.Errorf("expected container image %q, got %q", newImage.String(), image)
	}
	if containers[0].Name != expectedContainer {
		t.Errorf("expected container name %q, got %q", expectedContainer, containers[0].Name)
	}
}
