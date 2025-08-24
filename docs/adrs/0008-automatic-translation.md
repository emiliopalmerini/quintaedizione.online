# Automatic Translation 

Status:  Accepted

Context
- Tradurre a mano ogni collections è particolarmente oneroso
- utilizzare le api di Openai per tradurre i dati mancanti potrebbe essere più veloce

Decisione
- implementare un pulsante traduci in ../../editor/editor_app/templates/edit_raw.html che permetta di mandare il json del documento all'api di OpenAi
per tradurre i valori. 
- Dovremo creare una rotta api ../../editor/editor_app/routers/pages.py per chiamare tramite un openai_client
```python
client.responses.create(
    model="gpt-5",
    input="Write a one-sentence bedtime story about a unicorn."
)
```
- saranno da implementare tutte le tecniche necessarie 
https://platform.openai.com/docs/quickstart

Conseguenze
- Costo
- velocità

