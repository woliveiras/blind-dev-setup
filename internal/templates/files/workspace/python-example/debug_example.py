"""Exemplo sem dependências: breakpoint, variáveis e traceback intencional."""

import argparse


def average(values: list[int]) -> float:
    total = 0
    for value in values:
        total += value
    return total / len(values)


def main() -> None:
    parser = argparse.ArgumentParser(description="Exemplo de depuração do pendrive")
    parser.add_argument("--fail", action="store_true", help="provoca ZeroDivisionError")
    args = parser.parse_args()
    values = [] if args.fail else [2, 4, 6]
    result = average(values)
    print(f"Média: {result:.1f}")


if __name__ == "__main__":
    main()
