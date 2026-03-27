import argparse
import asyncio
import csv
import json
import time
import urllib.error
import urllib.request
from pathlib import Path


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="InferFlow starter load generator")
    parser.add_argument("--url", default="http://localhost:8080/v1/chat/completions")
    parser.add_argument("--requests", type=int, default=10)
    parser.add_argument("--model", default="mock-llm")
    parser.add_argument("--output", default="results/sample.csv")
    return parser.parse_args()


def build_payload(model: str, request_id: int) -> bytes:
    payload = {
        "model": model,
        "messages": [
            {
                "role": "user",
                "content": f"Starter request {request_id} from InferFlow loadgen",
            }
        ],
        "stream": False,
    }
    return json.dumps(payload).encode("utf-8")


async def issue_request(url: str, model: str, request_id: int) -> dict:
    started = time.time()
    prompt = f"Starter request {request_id} from InferFlow loadgen"

    def _send() -> dict:
        req = urllib.request.Request(
            url,
            data=build_payload(model, request_id),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(req, timeout=10) as resp:
            body = json.loads(resp.read().decode("utf-8"))
            return {"status": resp.status, "body": body}

    try:
        result = await asyncio.to_thread(_send)
        total_ms = int((time.time() - started) * 1000)
        text = result["body"]["choices"][0]["message"]["content"]
        return {
            "timestamp": int(started),
            "prompt_length": len(prompt),
            "strategy": "round_robin",
            "backend": result["body"].get("model", model),
            "ttft_ms": total_ms,
            "total_ms": total_ms,
            "tokens_generated": len(text.split()),
            "error": "",
        }
    except urllib.error.HTTPError as exc:
        total_ms = int((time.time() - started) * 1000)
        return {
            "timestamp": int(started),
            "prompt_length": len(prompt),
            "strategy": "round_robin",
            "backend": "",
            "ttft_ms": total_ms,
            "total_ms": total_ms,
            "tokens_generated": 0,
            "error": f"http_{exc.code}",
        }
    except Exception as exc:  # noqa: BLE001
        total_ms = int((time.time() - started) * 1000)
        return {
            "timestamp": int(started),
            "prompt_length": len(prompt),
            "strategy": "round_robin",
            "backend": "",
            "ttft_ms": total_ms,
            "total_ms": total_ms,
            "tokens_generated": 0,
            "error": str(exc),
        }


async def main() -> None:
    args = parse_args()
    rows = await asyncio.gather(
        *(issue_request(args.url, args.model, idx) for idx in range(args.requests))
    )

    output = Path(args.output)
    output.parent.mkdir(parents=True, exist_ok=True)

    with output.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.DictWriter(
            handle,
            fieldnames=[
                "timestamp",
                "prompt_length",
                "strategy",
                "backend",
                "ttft_ms",
                "total_ms",
                "tokens_generated",
                "error",
            ],
        )
        writer.writeheader()
        writer.writerows(rows)

    print(f"Wrote {len(rows)} rows to {output}")


if __name__ == "__main__":
    asyncio.run(main())
