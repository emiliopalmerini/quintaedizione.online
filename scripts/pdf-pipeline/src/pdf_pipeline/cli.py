"""CLI entrypoint for the PDF pipeline."""

from pathlib import Path

import click

from pdf_pipeline.convert import convert_pdf
from pdf_pipeline.split import split_markdown


@click.group()
def cli() -> None:
    """PDF-to-Markdown pipeline for D&D 5e SRD content."""


@cli.command()
@click.argument("pdf_path", type=click.Path(exists=True, path_type=Path))
@click.option(
    "--output-dir",
    type=click.Path(path_type=Path),
    default=None,
    help="Output directory for the converted markdown. Defaults to <pdf_dir>/converted/.",
)
def convert(pdf_path: Path, output_dir: Path | None) -> None:
    """Convert a PDF to a single markdown file with page markers."""
    if output_dir is None:
        output_dir = pdf_path.parent / "converted"

    click.echo(f"Converting {pdf_path} ...")
    result = convert_pdf(pdf_path, output_dir)
    click.echo(f"Done: {result}")


@cli.command()
@click.argument("md_path", type=click.Path(exists=True, path_type=Path))
@click.option(
    "--output-dir",
    type=click.Path(path_type=Path),
    default=None,
    help="Output directory for split files. Defaults to data/ita/lists/.",
)
def split(md_path: Path, output_dir: Path | None) -> None:
    """Split a converted markdown file into per-collection files."""
    if output_dir is None:
        output_dir = Path("data/ita/lists")

    click.echo(f"Splitting {md_path} ...")
    created = split_markdown(md_path, output_dir)
    click.echo(f"Done: {len(created)} files created.")


@cli.command()
@click.argument("pdf_path", type=click.Path(exists=True, path_type=Path))
@click.option(
    "--output-dir",
    type=click.Path(path_type=Path),
    default=None,
    help="Final output directory for per-collection files. Defaults to data/ita/lists/.",
)
def run(pdf_path: Path, output_dir: Path | None) -> None:
    """Run the full pipeline: convert PDF then split into collection files."""
    converted_dir = pdf_path.parent / "converted"

    click.echo(f"Step 1: Converting {pdf_path} ...")
    md_path = convert_pdf(pdf_path, converted_dir)
    click.echo(f"  Converted to: {md_path}")

    if output_dir is None:
        output_dir = Path("data/ita/lists")

    click.echo(f"Step 2: Splitting into {output_dir} ...")
    created = split_markdown(md_path, output_dir)
    click.echo(f"Done: {len(created)} files created.")
