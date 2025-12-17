package metadata

// Key identifies a CycloneDX field (or pseudo-field) we want to populate/check.
type Key string

func (k Key) String() string { return string(k) }

const (
	// BOM.metadata.component.* (MODEL)
	ComponentName               Key = "BOM.metadata.component.name"
	ComponentExternalReferences Key = "BOM.metadata.component.externalReferences"
	ComponentTags               Key = "BOM.metadata.component.tags"
	ComponentLicenses           Key = "BOM.metadata.component.licenses"
	ComponentHashes             Key = "BOM.metadata.component.hashes"
	ComponentManufacturer       Key = "BOM.metadata.component.manufacturer"
	ComponentAuthor             Key = "BOM.metadata.component.author"
	ComponentGroup              Key = "BOM.metadata.component.group"

	// Component-level extra properties (stored later as CycloneDX Component.Properties)
	ComponentPropertiesHuggingFaceLastModified Key = "BOM.metadata.component.properties.huggingface:lastModified"
	ComponentPropertiesHuggingFaceCreatedAt    Key = "BOM.metadata.component.properties.huggingface:createdAt"
	ComponentPropertiesHuggingFaceLanguage     Key = "BOM.metadata.component.properties.huggingface:language"
	ComponentPropertiesHuggingFaceUsedStorage  Key = "BOM.metadata.component.properties.huggingface:usedStorage"
	ComponentPropertiesHuggingFacePrivate      Key = "BOM.metadata.component.properties.huggingface:private"
	ComponentPropertiesHuggingFaceLibraryName  Key = "BOM.metadata.component.properties.huggingface:libraryName"
	ComponentPropertiesHuggingFaceDownloads    Key = "BOM.metadata.component.properties.huggingface:downloads"
	ComponentPropertiesHuggingFaceLikes        Key = "BOM.metadata.component.properties.huggingface:likes"

	// BOM.metadata.component.modelCard.* (MODEL CARD)
	ModelCardModelParametersTask               Key = "BOM.metadata.component.modelCard.modelParameters.task"
	ModelCardModelParametersArchitectureFamily Key = "BOM.metadata.component.modelCard.modelParameters.architectureFamily"
	ModelCardModelParametersModelArchitecture  Key = "BOM.metadata.component.modelCard.modelParameters.modelArchitecture"
	ModelCardModelParametersDatasets           Key = "BOM.metadata.component.modelCard.modelParameters.datasets"
)
