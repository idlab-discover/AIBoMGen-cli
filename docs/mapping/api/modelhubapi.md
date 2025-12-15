# Hub API

You can query a model from the hub API.

`cardData` contains the same yaml as the beginning of the modelCard.

GET https://huggingface.co/api/models/google-bert/bert-base-uncased

RESPONSE
```json
{
    "_id": "621ffdc036468d709f174338",
    "id": "google-bert/bert-base-uncased",              #BOM.metadata.component.name
    "private": false,
    "pipeline_tag": "fill-mask",                        #BOM.metadata.component.modelcard.modelparameters.task
    "library_name": "transformers",
    "tags": [                                           #BOM.metadata.component.tags
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
    "modelId": "google-bert/bert-base-uncased",         #BOM.metadata.component.name
    "author": "google-bert",                            #BOM.metadata.component.manufacturer, BOM.metadata.component.author, BOM.metadata.component.group
    "sha": "86b5e0934494bd15c9632b12f734a8a67f723594",  #BOM.metadata.component.hashes[] (hash of full huggingface repository)
    "lastModified": "2024-02-19T11:06:12.000Z",         #BOM.metadata.component.properties[]
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
        "architectures": [                              #BOM.metadata.component.modelcard.modelparameters.modelarchitecture
            "BertForMaskedLM"
        ],
        "model_type": "bert",                           #BOM.metadata.component.modelcard.modelparameters.architecturefamily
        "tokenizer_config": {}                          
    },
    "cardData": {
        "language": "en",                               #BOM.metadata.component.modelcard.properties
        "tags": [                                       #we already have tags I dont now why there is one also here
            "exbert"
        ],
        "license": "apache-2.0",                        #BOM.metadata.component.licenses
        "datasets": [                                   #BOM.metadata.component.modelcard.modelparameters.datastes.ref (ref to component with these names, which will later also be querried for rich dataset components)
            "bookcorpus",
            "wikipedia"
        ]
    },
    "transformersInfo": {                               #optionally add these as custom properties (but not usefull I think)
        "auto_model": "AutoModelForMaskedLM",
        "pipeline_tag": "fill-mask",
        "processor": "AutoTokenizer"
    },
    "siblings": [                                       #optionally add these as nested components in bom.metadata.component as type file or binary
                                                        # {
                                                        #   "type": "file",
                                                        #   "name": "pytorch_model.bin",
                                                        #   "hashes": [
                                                        #     { "alg": "SHA-256", "content": "..." }
                                                        #   ]
                                                        # }
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
    "createdAt": "2022-03-02T23:29:04.000Z",            #BOM.metadata.component.properties[]
    "safetensors": {                                    # Add this as properties on the Safetensors file component (if we define this as nested file component)
        "parameters": {
            "F32": 110106428
        },
        "total": 110106428
    },
    "inference": "warm",
    "usedStorage": 13397387509                          #BOM.metadata.component.properties[]
}
´´´