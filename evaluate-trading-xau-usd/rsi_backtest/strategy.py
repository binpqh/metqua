from __future__ import annotations

from dataclasses import dataclass
from typing import Optional


@dataclass(frozen=True)
class Scenario:
    name: str
    rsi_condition: str
    command: str
    stop_loss_pct: Optional[float] = None
    take_profit_pct: Optional[float] = None
    stop_loss_abs: Optional[float] = None
    take_profit_abs: Optional[float] = None


def condition_match(value: float, expression: str) -> bool:
    text = expression.replace(" ", "")

    if text.startswith("<"):
        return value < float(text[1:])
    if text.startswith(">"):
        return value > float(text[1:])
    if "-" in text:
        parts = text.split("-")
        if len(parts) != 2:
            raise ValueError(f"Invalid RSI range expression: {expression}")
        low = float(parts[0])
        high = float(parts[1])
        return low <= value <= high

    raise ValueError(f"Unsupported RSI expression: {expression}")