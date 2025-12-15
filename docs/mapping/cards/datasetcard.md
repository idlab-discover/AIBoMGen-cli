GET https://huggingface.co/datasets/templates/data-card-example/resolve/main/README.md

```yaml

---
language:
- {lang_0}
- {lang_1}
license: {license}                  # BOM.metadata.component.licenses
license_name: {license_name}
license_link: {license_link}
license_details: {license_details}
tags:                               # BOM.metadata.component.tags
- {tag_0}
- {tag_1}
- {tag_2}
- {tag_3}
annotations_creators:               # BOM.metadata.component.manufacturer, BOM.metadata.component.author, BOM.metadata.component.group
- {creator}
language_creators:
- {creator}
language_details:
- {bcp47_lang_0}
- {bcp47_lang_1}
pretty_name: {pretty_name}
size_categories:
- {number_of_elements_in_dataset}
source_datasets:
- {source_dataset_0}
- {source_dataset_1}
task_categories:
- {task_0}
- {task_1}
task_ids:
- {subtask_0}
- {subtask_1}
paperswithcode_id: {paperswithcode_id}
configs:                          
- config_name: {config_name_0}
  data_files:
  - split: {split_name_0}
    path: {file_path_0}   # BOM.metadata.component.data.contents.attachment (we can also hash each thingy)
  - split: {split_name_1}
    path: {file_path_1}
- config_name: {config_name_1}
  data_files:
  - split: {split_name_3}
    path: {file_path_3}

dataset_info:
  features:
    - name: {feature_name_0}
      dtype: {feature_dtype_0}
    - name: {feature_name_1}
      dtype: {feature_dtype_1}
    - name: {feature_name_2}
      dtype: {feature_dtype_2}
  config_name: {config_name}
  splits:
    - name: {split_name_0}
      num_bytes: {split_num_bytes_0}
      num_examples: {split_num_examples_0}
  download_size: {dataset_download_size}
  dataset_size: {dataset_size}

# It can also be a list of multiple subsets (also called "configurations"):
# ```yaml
# dataset_info:
#   - config_name: {config0}
#     features:
#       ...
#   - config_name: {config1}
#     features:
#       ...
# ```

extra_gated_fields:
- {field_name_0}: {field_type_0}
- {field_name_1}: {field_type_1}
- {field_name_2}: {field_type_2}
- {field_name_3}: {field_type_3}
extra_gated_prompt: {extra_gated_prompt}
---
```
```markdown
# Dataset Card for {{ pretty_name | default("Dataset Name", true) }}

<!-- Provide a quick summary of the dataset. -->

{{ dataset_summary | default("", true) }}

## Dataset Details

### Dataset Description

<!-- Provide a longer summary of what this dataset is. -->

{{ dataset_description | default("", true) }} // BOM.metadata.component.data.description

- **Curated by:** {{ curators | default("[More Information Needed]", true)}} // BOM.metadata.component.data.governance.stewards.organization.name // BOM.metadata.component.data.governance.custodians.organization.name
- **Funded by [optional]:** {{ funded_by | default("[More Information Needed]", true)}} // BOM.metadata.component.data.governance.owners.organization.name
- **Shared by [optional]:** {{ shared_by | default("[More Information Needed]", true)}} // BOM.metadata.component.data.governance.custodians.organization.name
- **Language(s) (NLP):** {{ language | default("[More Information Needed]", true)}}
- **License:** {{ license | default("[More Information Needed]", true)}} // BOM.metadata.component.license

### Dataset Sources [optional]

<!-- Provide the basic links for the dataset. -->

- **Repository:** {{ repo | default("[More Information Needed]", true)}}                                            // #BOM.metadata.component.externalRef (url to huggingface) or make it from model id
- **Paper [optional]:** {{ paper | default("[More Information Needed]", true)}}                                     // #BOM.metadata.component.externalRef (url to paper)
- **Demo [optional]:** {{ demo | default("[More Information Needed]", true)}}                                       // #BOM.metadata.component.externalRef (url to demo)

## Uses

<!-- Address questions around how the dataset is intended to be used. -->

### Direct Use

<!-- This section describes suitable use cases for the dataset. -->

{{ direct_use | default("[More Information Needed]", true)}}

### Out-of-Scope Use

<!-- This section addresses misuse, malicious use, and uses that the dataset will not work well for. -->                                             // #BOM.metadata.component.data.sensitive data

{{ out_of_scope_use | default("[More Information Needed]", true)}}

## Dataset Structure

<!-- This section provides a description of the dataset fields, and additional information about the dataset structure such as criteria used to create the splits, relationships between data points, etc. -->

{{ dataset_structure | default("[More Information Needed]", true)}}

## Dataset Creation

### Curation Rationale

<!-- Motivation for the creation of this dataset. -->

{{ curation_rationale_section | default("[More Information Needed]", true)}}

### Source Data

<!-- This section describes the source data (e.g. news text and headlines, social media posts, translated sentences, ...). -->

#### Data Collection and Processing

<!-- This section describes the data collection and processing process such as data selection criteria, filtering and normalization methods, tools and libraries used, etc. -->

{{ data_collection_and_processing_section | default("[More Information Needed]", true)}}

#### Who are the source data producers?

<!-- This section describes the people or systems who originally created the data. It should also include self-reported demographic or identity information for the source data creators if this information is available. -->

{{ source_data_producers_section | default("[More Information Needed]", true)}}

### Annotations [optional]

<!-- If the dataset contains annotations which are not part of the initial data collection, use this section to describe them. -->

#### Annotation process

<!-- This section describes the annotation process such as annotation tools used in the process, the amount of data annotated, annotation guidelines provided to the annotators, interannotator statistics, annotation validation, etc. -->

{{ annotation_process_section | default("[More Information Needed]", true)}}

#### Who are the annotators?

<!-- This section describes the people or systems who created the annotations. -->

{{ who_are_annotators_section | default("[More Information Needed]", true)}}

#### Personal and Sensitive Information

<!-- State whether the dataset contains data that might be considered personal, sensitive, or private (e.g., data that reveals addresses, uniquely identifiable names or aliases, racial or ethnic origins, sexual orientations, religious beliefs, political opinions, financial or health data, etc.). If efforts were made to anonymize the data, describe the anonymization process. --> 

{{ personal_and_sensitive_information | default("[More Information Needed]", true)}} // #BOM.metadata.component.data.sensitive data

## Bias, Risks, and Limitations

<!-- This section is meant to convey both technical and sociotechnical limitations. -->

{{ bias_risks_limitations | default("[More Information Needed]", true)}} // #BOM.metadata.component.data.sensitive data

### Recommendations

<!-- This section is meant to convey recommendations with respect to the bias, risk, and technical limitations. -->

{{ bias_recommendations | default("Users should be made aware of the risks, biases and limitations of the dataset. More information needed for further recommendations.", true)}}

## Citation [optional]

<!-- If there is a paper or blog post introducing the dataset, the APA and Bibtex information for that should go in this section. -->

**BibTeX:**

{{ citation_bibtex | default("[More Information Needed]", true)}}

**APA:**

{{ citation_apa | default("[More Information Needed]", true)}}

## Glossary [optional]

<!-- If relevant, include terms and calculations in this section that can help readers understand the dataset or dataset card. -->

{{ glossary | default("[More Information Needed]", true)}}

## More Information [optional]

{{ more_information | default("[More Information Needed]", true)}}

## Dataset Card Authors [optional]

{{ dataset_card_authors | default("[More Information Needed]", true)}}

## Dataset Card Contact

{{ dataset_card_contact | default("[More Information Needed]", true)}}
// BOM.metadata.component.properties (datasetcardcontact)

```