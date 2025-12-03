# Huggingface metadata sources:

## 1. HF_API /api/models/:id (https://huggingface.co/api/models/:id)

GET https://huggingface.co/api/models/google-bert/bert-base-uncased

RESPONSE
`
{
    "_id": "621ffdc036468d709f174338",
    "id": "google-bert/bert-base-uncased",
    "private": false,
    "pipeline_tag": "fill-mask",
    "library_name": "transformers",
    "tags": [
        "transformers",
        "pytorch",
        "tf",
        "jax",
        "rust",
        "coreml",
        "onnx",
        "safetensors",
        "bert",
        "fill-mask",
        "exbert",
        "en",
        "dataset:bookcorpus",
        "dataset:wikipedia",
        "arxiv:1810.04805",
        "license:apache-2.0",
        "endpoints_compatible",
        "deploy:azure",
        "region:us"
    ],
    "downloads": 60478923,
    "likes": 2498,
    "modelId": "google-bert/bert-base-uncased",
    "author": "google-bert",
    "sha": "86b5e0934494bd15c9632b12f734a8a67f723594",
    "lastModified": "2024-02-19T11:06:12.000Z",
    "gated": false,
    "disabled": false,
    "mask_token": "[MASK]",
    "widgetData": [
        {
            "text": "Paris is the [MASK] of France."
        },
        {
            "text": "The goal of life is [MASK]."
        }
    ],
    "model-index": null,
    "config": {
        "architectures": [
            "BertForMaskedLM"
        ],
        "model_type": "bert",
        "tokenizer_config": {}
    },
    "cardData": {
        "language": "en",
        "tags": [
            "exbert"
        ],
        "license": "apache-2.0",
        "datasets": [
            "bookcorpus",
            "wikipedia"
        ]
    },
    "transformersInfo": {
        "auto_model": "AutoModelForMaskedLM",
        "pipeline_tag": "fill-mask",
        "processor": "AutoTokenizer"
    },
    "siblings": [
        {
            "rfilename": ".gitattributes"
        },
        {
            "rfilename": "LICENSE"
        },
        {
            "rfilename": "README.md"
        },
        {
            "rfilename": "config.json"
        },
        {
            "rfilename": "coreml/fill-mask/float32_model.mlpackage/Data/com.apple.CoreML/model.mlmodel"
        },
        {
            "rfilename": "coreml/fill-mask/float32_model.mlpackage/Data/com.apple.CoreML/weights/weight.bin"
        },
        {
            "rfilename": "coreml/fill-mask/float32_model.mlpackage/Manifest.json"
        },
        {
            "rfilename": "flax_model.msgpack"
        },
        {
            "rfilename": "model.onnx"
        },
        {
            "rfilename": "model.safetensors"
        },
        {
            "rfilename": "pytorch_model.bin"
        },
        {
            "rfilename": "rust_model.ot"
        },
        {
            "rfilename": "tf_model.h5"
        },
        {
            "rfilename": "tokenizer.json"
        },
        {
            "rfilename": "tokenizer_config.json"
        },
        {
            "rfilename": "vocab.txt"
        }
    ],
    "spaces": [
        "jiani-huang/LASER",
        "nikigoli/countgd",
        "exbert-project/exbert",
        "merve/Grounding_DINO_demo",
        "Yiwen-ntu/MeshAnythingV2",
        "kevinwang676/E2-F5-TTS",
        "ChaolongYang/KDTalker",
        "MCP-1st-Birthday/ML-Starter",
        "OpenEvals/tokenizers-languages",
        "mugu5/DL_Project",
        "microsoft/HuggingGPT",
        "Vision-CAIR/minigpt4",
        "lnyan/stablediffusion-infinity",
        "multimodalart/latentdiffusion",
        "mrfakename/MeloTTS",
        "Salesforce/BLIP",
        "shi-labs/Versatile-Diffusion",
        "yizhangliu/Grounded-Segment-Anything",
        "stepfun-ai/Step1X-Edit",
        "H-Liu1997/TANGO",
        "alexnasa/Chain-of-Zoom",
        "xinyu1205/recognize-anything",
        "hilamanor/audioEditing",
        "cvlab/zero123-live",
        "AIGC-Audio/AudioGPT",
        "m-ric/chunk_visualizer",
        "Audio-AGI/AudioSep",
        "jadechoghari/OpenMusic",
        "DAMO-NLP-SG/Video-LLaMA",
        "gligen/demo",
        "declare-lab/mustango",
        "Yiwen-ntu/MeshAnything",
        "shgao/EditAnything",
        "LiruiZhao/Diffree",
        "multimodalart/MoDA-fast-talking-head",
        "Vision-CAIR/MiniGPT-v2",
        "Yuliang/ECON",
        "THUdyh/Oryx",
        "IDEA-Research/Grounded-SAM",
        "OpenSound/CapSpeech-TTS",
        "Awiny/Image2Paragraph",
        "ShilongLiu/Grounding_DINO_demo",
        "yangheng/Super-Resolution-Anime-Diffusion",
        "liuyuan-pal/SyncDreamer",
        "XiangJinYu/SPO",
        "sam-hq-team/sam-hq",
        "haotiz/glip-zeroshot-demo",
        "Nick088/Audio-SR",
        "TencentARC/BrushEdit",
        "nateraw/lavila",
        "abyildirim/inst-inpaint",
        "Pinwheel/GLIP-BLIP-Object-Detection-VQA",
        "Junfeng5/GLEE_demo",
        "shi-labs/Matting-Anything",
        "fffiloni/Video-Matting-Anything",
        "burtenshaw/autotrain-mcp",
        "Vision-CAIR/MiniGPT4-video",
        "linfanluntan/Grounded-SAM",
        "magicr/BuboGPT",
        "WensongSong/Insert-Anything",
        "mteb/leaderboard_legacy",
        "nvidia/audio-flamingo-2",
        "clip-italian/clip-italian-demo",
        "OpenGVLab/InternGPT",
        "3DTopia/3DTopia",
        "yenniejun/tokenizers-languages",
        "mmlab-ntu/relate-anything-model",
        "amphion/PicoAudio",
        "byeongjun-park/HarmonyView",
        "fffiloni/vta-ldm",
        "keras-io/bert-semantic-similarity",
        "MirageML/sjc",
        "zakaria-narjis/photo-enhancer",
        "NAACL2022/CLIP-Caption-Reward",
        "society-ethics/model-card-regulatory-check",
        "fffiloni/miniGPT4-Video-Zero",
        "AIGC-Audio/AudioLCM",
        "Gladiator/Text-Summarizer",
        "SVGRender/DiffSketcher",
        "ethanchern/Anole",
        "topdu/OpenOCR-Demo",
        "acmc/whatsapp-chats-finetuning-formatter",
        "LittleFrog/IntrinsicAnything",
        "milyiyo/reimagine-it",
        "ysharma/text-to-image-to-video",
        "OpenGVLab/VideoChatGPT",
        "sonalkum/GAMA",
        "kaushalya/medclip-roco",
        "AIGC-Audio/Make_An_Audio",
        "avid-ml/bias-detection",
        "ZebangCheng/Emotion-LLaMA",
        "bartar/tokenizers",
        "RitaParadaRamos/SmallCapDemo",
        "llizhx/TinyGPT-V",
        "codelion/Grounding_DINO_demo",
        "flosstradamus/FluxMusicGUI",
        "Tinkering/Pytorch-day-prez",
        "sasha/BiasDetection",
        "Pusheen/LoCo",
        "Jingkang/EgoGPT-7B"
    ],
    "createdAt": "2022-03-02T23:29:04.000Z",
    "safetensors": {
        "parameters": {
            "F32": 110106428
        },
        "total": 110106428
    },
    "inference": "warm",
    "usedStorage": 13397387509
}
`

## 2. README_TEXT

Regex-based extraction from README.

cardData field of previous HF_API equals to the yaml at the start of this readme file (following example has empty yaml and thuss cardData)

The template of such a README modelCard looks like this:

GET https://huggingface.co/templates/model-card-example/resolve/main/README.md

RESPONSE
```markdown
---
# For reference on model card metadata, see the spec: https://github.com/huggingface/hub-docs/blob/main/modelcard.md?plain=1
# Doc / guide: https://huggingface.co/docs/hub/model-cards
{}
---

# Model Card for Model ID

<!-- Provide a quick summary of what the model is/does. -->

This modelcard aims to be a base template for new models. It has been generated using [this raw template](https://github.com/huggingface/huggingface_hub/blob/main/src/huggingface_hub/templates/modelcard_template.md?plain=1).

## Model Details

### Model Description

<!-- Provide a longer summary of what this model is. -->



- **Developed by:** [More Information Needed]
- **Funded by [optional]:** [More Information Needed]
- **Shared by [optional]:** [More Information Needed]
- **Model type:** [More Information Needed]
- **Language(s) (NLP):** [More Information Needed]
- **License:** [More Information Needed]
- **Finetuned from model [optional]:** [More Information Needed]

### Model Sources [optional]

<!-- Provide the basic links for the model. -->

- **Repository:** [More Information Needed]
- **Paper [optional]:** [More Information Needed]
- **Demo [optional]:** [More Information Needed]

## Uses

<!-- Address questions around how the model is intended to be used, including the foreseeable users of the model and those affected by the model. -->

### Direct Use

<!-- This section is for the model use without fine-tuning or plugging into a larger ecosystem/app. -->

[More Information Needed]

### Downstream Use [optional]

<!-- This section is for the model use when fine-tuned for a task, or when plugged into a larger ecosystem/app -->

[More Information Needed]

### Out-of-Scope Use

<!-- This section addresses misuse, malicious use, and uses that the model will not work well for. -->

[More Information Needed]

## Bias, Risks, and Limitations

<!-- This section is meant to convey both technical and sociotechnical limitations. -->

[More Information Needed]

### Recommendations

<!-- This section is meant to convey recommendations with respect to the bias, risk, and technical limitations. -->

Users (both direct and downstream) should be made aware of the risks, biases and limitations of the model. More information needed for further recommendations.

## How to Get Started with the Model

Use the code below to get started with the model.

[More Information Needed]

## Training Details

### Training Data

<!-- This should link to a Dataset Card, perhaps with a short stub of information on what the training data is all about as well as documentation related to data pre-processing or additional filtering. -->

[More Information Needed]

### Training Procedure

<!-- This relates heavily to the Technical Specifications. Content here should link to that section when it is relevant to the training procedure. -->

#### Preprocessing [optional]

[More Information Needed]


#### Training Hyperparameters

- **Training regime:** [More Information Needed] <!--fp32, fp16 mixed precision, bf16 mixed precision, bf16 non-mixed precision, fp16 non-mixed precision, fp8 mixed precision -->

#### Speeds, Sizes, Times [optional]

<!-- This section provides information about throughput, start/end time, checkpoint size if relevant, etc. -->

[More Information Needed]

## Evaluation

<!-- This section describes the evaluation protocols and provides the results. -->

### Testing Data, Factors & Metrics

#### Testing Data

<!-- This should link to a Dataset Card if possible. -->

[More Information Needed]

#### Factors

<!-- These are the things the evaluation is disaggregating by, e.g., subpopulations or domains. -->

[More Information Needed]

#### Metrics

<!-- These are the evaluation metrics being used, ideally with a description of why. -->

[More Information Needed]

### Results

[More Information Needed]

#### Summary



## Model Examination [optional]

<!-- Relevant interpretability work for the model goes here -->

[More Information Needed]

## Environmental Impact

<!-- Total emissions (in grams of CO2eq) and additional considerations, such as electricity usage, go here. Edit the suggested text below accordingly -->

Carbon emissions can be estimated using the [Machine Learning Impact calculator](https://mlco2.github.io/impact#compute) presented in [Lacoste et al. (2019)](https://arxiv.org/abs/1910.09700).

- **Hardware Type:** [More Information Needed]
- **Hours used:** [More Information Needed]
- **Cloud Provider:** [More Information Needed]
- **Compute Region:** [More Information Needed]
- **Carbon Emitted:** [More Information Needed]

## Technical Specifications [optional]

### Model Architecture and Objective

[More Information Needed]

### Compute Infrastructure

[More Information Needed]

#### Hardware

[More Information Needed]

#### Software

[More Information Needed]

## Citation [optional]

<!-- If there is a paper or blog post introducing the model, the APA and Bibtex information for that should go in this section. -->

**BibTeX:**

[More Information Needed]

**APA:**

[More Information Needed]

## Glossary [optional]

<!-- If relevant, include terms and calculations in this section that can help readers understand the model or model card. -->

[More Information Needed]

## More Information [optional]

[More Information Needed]

## Model Card Authors [optional]

[More Information Needed]

## Model Card Contact

[More Information Needed]
```

## 3. CONFIG_FILE

GET https://huggingface.co/google-bert/bert-base-uncased/resolve/main/config.json

RESPONSE
```json
{
  "architectures": [
    "BertForMaskedLM"
  ],
  "attention_probs_dropout_prob": 0.1,
  "gradient_checkpointing": false,
  "hidden_act": "gelu",
  "hidden_dropout_prob": 0.1,
  "hidden_size": 768,
  "initializer_range": 0.02,
  "intermediate_size": 3072,
  "layer_norm_eps": 1e-12,
  "max_position_embeddings": 512,
  "model_type": "bert",
  "num_attention_heads": 12,
  "num_hidden_layers": 12,
  "pad_token_id": 0,
  "position_embedding_type": "absolute",
  "transformers_version": "4.6.0.dev0",
  "type_vocab_size": 2,
  "use_cache": true,
  "vocab_size": 30522
}

```

## 4. REPOSITORY_FILES

GET https://huggingface.co/google-bert/bert-base-uncased/resolve/main/:file

file can be: 
- LICENSE
- README.md (see previous)
- config.json (see previous)
- model.onnx
- model.safetensors...
- pytorch_model.bin
- rust_model.ot
- tf_model.h5
- tokenizer.json
- tokenizer_config.json
- vocab.txt

# Huggingface dataset metadata sources:

you can find model datasets in cardData (the yaml of the readme or via the hub API (best) )
or you can find them in the tags (via the hub API)

then you can query also these same way to find out information about these datasets

