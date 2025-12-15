

# Hub API

You can query a dataset from the hub API.

`cardData` contains the same yaml as the beginning of the datasetcard.

GET https://huggingface.co/api/datasets/LanguageShades/BiasShades

RESPONSE
```json
{
    "_id": "666ae9a84649c647cfb85a54",
    "id": "LanguageShades/BiasShades",
    "author": "LanguageShades",                            # BOM.metadata.component.data.governance.custodians.organization.name
    "sha": "dc956d4bed06c89e5e49dbeeace69b039377e007",     #BOM.metadata.component.hashes[] (hash of full huggingface repository)
    "lastModified": "2025-05-03T23:25:04.000Z",    # BOM.metadata.component.tags
    "private": false,
    "gated": "auto",
    "disabled": false,
    "tags": [
        "task_categories:text-classification",   # BOM.metadata.component.tags
        "language:ar",
        "language:bn",
        "language:de",
        "language:en",
        "language:es",
        "language:hi",
        "language:it",
        "language:mr",
        "language:nl",
        "language:pl",
        "language:ro",
        "language:ru",
        "language:zh",
        "language:pt",
        "size_categories:n<1K",
        "format:csv",
        "modality:image",
        "modality:tabular",
        "modality:text",
        "library:datasets",
        "library:pandas",
        "library:mlcroissant",
        "library:polars",
        "region:us",
        "stereotype",
        "social bias",
        "socialbias"
    ],
    "citation": null,
    "description": "Interested in contributing? Speak a language not represented here? Disagree with an annotation? Please submit feedback in the Community tab!\n\n\t\n\t\t\n\t\tDataset Card for BiasShades\n\t\n\nNote: This dataset may NOT be used as training data in any form (pre-training, fine-tuning, post-training, etc.) without express permission from creators.\n\n\t\n\t\t\n\t\tDataset Details\n\t\n\nVersion: Beta \nInitial dataset for public launch on day of NAACL presentation. Minor changes and License to follow.\n\n\t\n\t\t\n\t\tDataset… See the full description on the dataset page: https://huggingface.co/datasets/LanguageShades/BiasShades.",           # BOM.metadata.component.data.description
    "downloads": 106,
    "likes": 23,
    "cardData": {
        "task_categories": [            # BOM.metadata.component.data.classification      
            "text-classification",
            "text2text-generation"
        ],
        "language": [
            "ar",
            "bn",
            "de",
            "en",
            "es",
            "hi",
            "it",
            "mr",
            "nl",
            "pl",
            "ro",
            "ru",
            "zh",
            "pt"
        ],
        "configs": [
            {
                "config_name": "by_language",
                "data_files": [
                    {
                        "split": "ar",
                        "path": "by_language/ar.csv"
                    },
                    {
                        "split": "bn",
                        "path": "by_language/bn.csv"
                    },
                    {
                        "split": "de",
                        "path": "by_language/de.csv"
                    },
                    {
                        "split": "en",
                        "path": "by_language/en.csv"
                    },
                    {
                        "split": "es",
                        "path": "by_language/es.csv"
                    },
                    {
                        "split": "fr",
                        "path": "by_language/fr.csv"
                    },
                    {
                        "split": "hi",
                        "path": "by_language/hi.csv"
                    },
                    {
                        "split": "it",
                        "path": "by_language/it.csv"
                    },
                    {
                        "split": "mr",
                        "path": "by_language/mr.csv"
                    },
                    {
                        "split": "nl",
                        "path": "by_language/nl.csv"
                    },
                    {
                        "split": "pl",
                        "path": "by_language/pl.csv"
                    },
                    {
                        "split": "pt_br",
                        "path": "by_language/pt_br.csv"
                    },
                    {
                        "split": "ro",
                        "path": "by_language/ro.csv"
                    },
                    {
                        "split": "ru",
                        "path": "by_language/ru.csv"
                    },
                    {
                        "split": "zh",
                        "path": "by_language/zh.csv"
                    },
                    {
                        "split": "zh_hant",
                        "path": "by_language/zh_hant.csv"
                    }
                ]
            },
            {
                "config_name": "default",
                "data_files": [
                    {
                        "split": "test",
                        "path": "all/all.csv"
                    }
                ]
            }
        ],
        "tags": [               # BOM.metadata.component.data.sensitiveData []
            "stereotype",
            "social bias",
            "socialbias"
        ],
        "size_categories": [
            "n<1K"
        ],
        "extra_gated_prompt": "You must agree not to use the dataset as training data.",
        "extra_gated_fields": {
            "I want to use this dataset for": {
                "type": "select",
                "options": [
                    "Research",
                    "Education",
                    {
                        "label": "Other",
                        "value": "other"
                    }
                ]
            },
            "I agree not to use this dataset for training": "checkbox"
        }
    },
    "siblings": [
        {
            "rfilename": ".gitattributes"
        },
        {
            "rfilename": "README.md"
        },
        {
            "rfilename": "all/all.csv"
        },
        {
            "rfilename": "average_number_of_regions_recognizing_stereotype_barchart.png"
        },
        {
            "rfilename": "bias_type.png"
        },
        {
            "rfilename": "bias_type_by_region.png"
        },
        {
            "rfilename": "bias_type_stereotyped_entities_sunburst.png"
        },
        {
            "rfilename": "by_language/ar.csv"
        },
        {
            "rfilename": "by_language/bn.csv"
        },
        {
            "rfilename": "by_language/de.csv"
        },
        {
            "rfilename": "by_language/en.csv"
        },
        {
            "rfilename": "by_language/es.csv"
        },
        {
            "rfilename": "by_language/fr.csv"
        },
        {
            "rfilename": "by_language/hi.csv"
        },
        {
            "rfilename": "by_language/it.csv"
        },
        {
            "rfilename": "by_language/mr.csv"
        },
        {
            "rfilename": "by_language/nl.csv"
        },
        {
            "rfilename": "by_language/pl.csv"
        },
        {
            "rfilename": "by_language/pt_br.csv"
        },
        {
            "rfilename": "by_language/ro.csv"
        },
        {
            "rfilename": "by_language/ru.csv"
        },
        {
            "rfilename": "by_language/zh.csv"
        },
        {
            "rfilename": "by_language/zh_hant.csv"
        },
        {
            "rfilename": "creator_ages.png"
        },
        {
            "rfilename": "creator_background_socioec_classes.png"
        },
        {
            "rfilename": "creator_birth_countries.png"
        },
        {
            "rfilename": "creator_country_residences.png"
        },
        {
            "rfilename": "creator_current_socioec_classes.png"
        },
        {
            "rfilename": "creator_degree_status.png"
        },
        {
            "rfilename": "creator_genders.png"
        },
        {
            "rfilename": "creator_languages.png"
        },
        {
            "rfilename": "creator_native_languages.png"
        },
        {
            "rfilename": "creator_primary_occupations.png"
        },
        {
            "rfilename": "distribution_of_bias_types_across_all_languages_and_regions_barchart.png"
        },
        {
            "rfilename": "distribution_of_bias_types_across_all_languages_and_regions_pie.png"
        },
        {
            "rfilename": "distribution_of_bias_types_and_stereotyped_entities_sunburst.png"
        },
        {
            "rfilename": "distribution_of_bias_types_and_stereotyped_entities_sunburst_combined.png"
        },
        {
            "rfilename": "distribution_of_bias_types_by_statement_form_sunburst.png"
        },
        {
            "rfilename": "distribution_of_recognized_bias_types_across_all_languages_and_regions_pie_camera-ready.png"
        },
        {
            "rfilename": "distribution_of_recognized_stereotypes_across_all_languages_and_regions_barchart.png"
        },
        {
            "rfilename": "distribution_of_recognized_stereotypes_across_regions_by_bias_type_bar.png"
        },
        {
            "rfilename": "distribution_of_recognized_stereotypes_across_regions_by_bias_type_sunburst.png"
        },
        {
            "rfilename": "shades.csv"
        },
        {
            "rfilename": "shades_map.pdf"
        }
    ],
    "createdAt": "2024-06-13T12:44:24.000Z",    # BOM.metadata.component.properties
    "usedStorage": 7258598                      # BOM.metadata.component.properties
}
```