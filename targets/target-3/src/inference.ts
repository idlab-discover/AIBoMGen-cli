/**
 * Inference using @huggingface/transformers (browser/Node.js).
 *
 * Detectable patterns:
 *   - pipeline(<task>, <org/model>)  positional
 *   - .from_pretrained(<org/model>)
 *   - model: <org/model>  in object literals / function args
 *
 * NOT detectable:
 *   - Variable holding model ID
 *   - Template literals: `${org}/${model}`
 */

import { pipeline, AutoModel, AutoTokenizer, env } from "@huggingface/transformers";
import { HfInference } from "@huggingface/inference";

// ── pipeline() ───────────────────────────────────────────────────────────────

// Single-line positional form – detected by js_pipeline_positional
const classifier = await pipeline("text-classification", "distilbert/distilbert-base-uncased-finetuned-sst-2-english");

// With options object on same line – also detected
const generator = await pipeline("text-generation", "Xenova/gpt2", { dtype: "q4" });

// NOT detectable: model ID split onto a second line (scanner does not stitch JS):
// const gen2 = await pipeline(
//   <text-generation>,
//   <meta-llama/Llama-3.1-8B-Instruct>  // would be missed
// );

// ── .from_pretrained() ───────────────────────────────────────────────────────

// js_from_pretrained rule
const model = await AutoModel.from_pretrained("Xenova/bert-base-uncased");
const tokenizer = await AutoTokenizer.from_pretrained("Xenova/bert-base-uncased");

// Quantised model
const qModel = await AutoModel.from_pretrained(
  "Xenova/distilbert-base-uncased-finetuned-sst-2-english"
);

// ── @huggingface/inference ────────────────────────────────────────────────────

const hf = new HfInference(process.env.HF_TOKEN);

// model: field inside function argument object – js_model_field rule (requires org/model)
const out1 = await hf.textGeneration({
  model: "meta-llama/Llama-3.1-8B-Instruct",
  inputs: "Once upon a time",
});

const out2 = await hf.imageClassification({
  model: "google/vit-base-patch16-224",
  data: imageBlob,
});

// ── NOT detectable: variable indirection ────────────────────────────────────

const MODEL = "facebook/opt-1.3b";
const dynamicModel = await AutoModel.from_pretrained(MODEL); // variable → not caught

// NOT detectable: template literal
const org = "Xenova";
const name = "bert-base-uncased";
const tlModel = await AutoModel.from_pretrained(`${org}/${name}`);

// NOT detectable: ternary choosing between two IDs stored in variables
const useSmall = true;
const chosenModel = useSmall ? "Xenova/gpt2" : "Xenova/llama-3-8b";
const ternaryModel = await AutoModel.from_pretrained(chosenModel);
