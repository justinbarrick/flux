package resource

import (
	"fmt"
	"sort"

	"github.com/weaveworks/flux/image"
	"github.com/weaveworks/flux/resource"
)

// ReleaseContainerName is the name used when flux interprets a
// FluxHelmRelease as having a container with an image, by virtue of
// having a `values` stanza with an image field:
//
// spec:
//   ...
//   values:
//     image: some/image:version
//
// The name refers to the source of the image value.
const ReleaseContainerName = "chart-image"

// FluxHelmRelease echoes the generated type for the custom resource
// definition. It's here so we can 1. get `baseObject` in there, and
// 3. control the YAML serialisation of fields, which we can't do
// (easily?) with the generated type.
type FluxHelmRelease struct {
	baseObject
	Spec struct {
		Values map[string]interface{}
	}
}

type ImageSetter func(image.Ref)

// The type we have to interpret as containers is a
// `map[string]interface{}`; and, we want a stable order to the
// containers we output, since things will jump around in API calls,
// or fail to verify, otherwise. Since we can't get them in the order
// they appear in the document, sort them.
func sorted_keys(values map[string]interface{}) []string {
	var keys []string
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// FindFluxHelmReleaseContainers examines the Values from a
// FluxHelmRelease (manifest, or cluster resource, or otherwise) and
// calls visit with each container name and image it finds, as well as
// procedure for changing the image value. It will return an error if
// it cannot interpret the values as specifying images, or if the
// `visit` function itself returns an error.
func FindFluxHelmReleaseContainers(values map[string]interface{}, visit func(string, image.Ref, ImageSetter) error) error {
	// Try the simplest format first:
	// ```
	// values:
	//   image: 'repo/image:tag'
	// ```
	if image, setter, ok := interpret_stringmap(values); ok {
		visit(ReleaseContainerName, image, setter)
		return nil
	}

	// Second most simple format:
	// ```
	// values:
	//   foo:
	//     image: repo/foo:v1
	//   bar:
	//     image: repo/bar:v2
	// ```
	// with the variation that there may also be a `tag` field:
	// ```
	// values:
	//   foo:
	//     image: repo/foo
	//     tag: v1
	for _, k := range sorted_keys(values) {
		// From a YAML (i.e., a file), it's a
		// `map[interface{}]interface{}`, and from JSON (i.e.,
		// Kubernetes API) it's a `map[string]interface{}`.
		switch m := values[k].(type) {
		case map[string]interface{}:
			if image, setter, ok := interpret_stringmap(m); ok {
				visit(k, image, setter)
			}
		case map[interface{}]interface{}:
			if image, setter, ok := interpret_anymap(m); ok {
				visit(k, image, setter)
			}
		}
	}
	return nil
}

func interpret_stringmap(m map[string]interface{}) (image.Ref, ImageSetter, bool) {
	switch img := m["image"].(type) {
	case string:
		imageRef, err := image.ParseRef(img)
		if err == nil {
			var taggy bool
			if tag, ok := m["tag"]; ok {
				if tagStr, ok := tag.(string); ok {
					taggy = true
					imageRef.Tag = tagStr
				}
			}
			return imageRef, func(ref image.Ref) {
				if taggy {
					m["image"] = ref.Name.String()
					m["tag"] = ref.Tag
					return
				}
				m["image"] = ref.String()
			}, true
		}
	case map[string]interface{}:
		if imgRepo, ok := img["repository"].(string); ok {
			if imgTag, ok := img["tag"].(string); ok {
				imgRef, err := image.ParseRef(imgRepo + ":" + imgTag)
				if err == nil {
					return imgRef, func(ref image.Ref) {
						img["repository"] = ref.Name.String()
						img["tag"] = ref.Tag
					}, true
				}
			}
		}
	case map[interface{}]interface{}:
		if imgRepo, ok := img["repository"].(string); ok {
			if imgTag, ok := img["tag"].(string); ok {
				imgRef, err := image.ParseRef(imgRepo + ":" + imgTag)
				if err == nil {
					return imgRef, func(ref image.Ref) {
						img["repository"] = ref.Name.String()
						img["tag"] = ref.Tag
					}, true
				}
			}
		}
	}
	return image.Ref{}, nil, false
}

// Almost exactly the same code, lexically. just a different type, because go.
func interpret_anymap(m map[interface{}]interface{}) (image.Ref, ImageSetter, bool) {
	switch img := m["image"].(type) {
	case string:
		imageRef, err := image.ParseRef(img)
		if err == nil {
			var taggy bool
			if tag, ok := m["tag"]; ok {
				if tagStr, ok := tag.(string); ok {
					taggy = true
					imageRef.Tag = tagStr
				}
			}
			return imageRef, func(ref image.Ref) {
				if taggy {
					m["image"] = ref.Name.String()
					m["tag"] = ref.Tag
					return
				}
				m["image"] = ref.String()
			}, true
		}
	case map[interface{}]interface{}:
		if imgRepo, ok := img["repository"].(string); ok {
			if imgTag, ok := img["tag"].(string); ok {
				imgRef, err := image.ParseRef(imgRepo + ":" + imgTag)
				if err == nil {
					return imgRef, func(ref image.Ref) {
						img["repository"] = ref.Name.String()
						img["tag"] = ref.Tag
					}, true
				}
			}
		}
	}
	return image.Ref{}, nil, false
}

// Containers returns the containers that are defined in the
// FluxHelmRelease.
func (fhr FluxHelmRelease) Containers() []resource.Container {
	var containers []resource.Container
	// If there's an error in interpreting, return what we have.
	_ = FindFluxHelmReleaseContainers(fhr.Spec.Values, func(container string, image image.Ref, _ ImageSetter) error {
		containers = append(containers, resource.Container{
			Name:  container,
			Image: image,
		})
		return nil
	})
	return containers
}

// SetContainerImage mutates this resource by setting the `image`
// field of `values`, or a subvalue therein, per one of the
// interpretations in `FindFluxHelmReleaseContainers` above. NB we can
// get away with a value-typed receiver because we set a map entry.
func (fhr FluxHelmRelease) SetContainerImage(container string, ref image.Ref) error {
	found := false
	if err := FindFluxHelmReleaseContainers(fhr.Spec.Values, func(name string, image image.Ref, setter ImageSetter) error {
		if container == name {
			setter(ref)
			found = true
		}
		return nil
	}); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("did not find container %s in FluxHelmRelease", container)
	}
	return nil
}
