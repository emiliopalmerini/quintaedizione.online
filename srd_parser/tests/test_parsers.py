from srd_parser.parsers.classes_improved import parse_classes
from srd_parser.parsers.backgrounds import parse_backgrounds


def test_parse_classes_minimal():
    md = [
        "# Classi",
        "",
        "## Guerriero",
        "Tabella: Tratti base del Guerriero",
        "| Etichetta | Valore |",
        "|---|---|",
        "| Dado Punti Ferita | d10 |",
        "",
        "Tabella: Privilegi del Guerriero",
        "| Livello | Bonus competenza | Privilegi di classe |",
        "|---|---|---|",
        "| 1 | +2 | Talento |",
    ]
    docs = parse_classes(md)
    assert any(d.get("nome") == "Guerriero" for d in docs)


def test_parse_backgrounds_minimal():
    md = [
        "# Origini del Personaggio",
        "",
        "### Descrizioni dei Background",
        "#### Accolito",
        "**Punteggi di Caratteristica:** Intelligenza, Saggezza",
        "**Talento:** Iniziato alla Magia (Chierico)",
        "**Competenze in Abilità:** Intuizione e Religione",
        "**Competenza negli Strumenti:** Strumenti da Calligrafo",
        "**Equipaggiamento:** (A) Oggetto; oppure (B) Oggetto",
    ]
    docs = parse_backgrounds(md)
    assert any(d.get("nome") == "Accolito" for d in docs)

