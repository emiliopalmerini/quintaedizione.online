#!/usr/bin/env python3
"""Translate and condense map descriptions to Italian.

Reads descriptions_raw.json and mappe_curated.json,
applies Italian translations, and merges into data/mappe/mappe.json.

Usage:
    python scripts/scrape_maps/translate_descriptions.py
"""

import json
from pathlib import Path

DATA_DIR = Path(__file__).resolve().parents[2] / "data" / "mappe"

# Map source_url → Italian description (1-2 evocative sentences)
DESCRIPTIONS = {
    "https://dysonlogos.blog/2025/08/13/the-dreadwarren/":
        "Costruito per onorare la nobile casata Virell, il Dreadwarren fu abbandonato dopo una pestilenza. Ora banditi occupano le cripte orientali, mentre un giovane negromante ha reclamato il sanctum occidentale.",

    "https://dysonlogos.blog/2025/07/16/hexatomb/":
        "Un antico sepolcreto esagonale ai confini delle terre coltivate, fonte di storie di fantasmi. All'interno si trovano segni di feste clandestine, un ricovero per capre e persino un vecchio campo di tiro con l'arco.",

    "https://dysonlogos.blog/2025/06/20/shrine-on-the-mosswater/":
        "Un piccolo santuario sorge lungo il corso del Mosswater, circondato da muschio e pietre antiche. Un luogo misterioso dove il confine tra il sacro e il selvaggio si assottiglia.",

    "https://dysonlogos.blog/2025/08/11/index-card-dungeon-ii-map-12-deep-ruins/":
        "Rovine profonde sotto la Torre, raggiungibili tramite una porta segreta. I maestri del Culto dell'Occhio Fratturato sono partiti per esplorare questi passaggi diversi giorni fa.",

    "https://dysonlogos.blog/2025/08/08/index-card-dungeon-ii-map-11-ancient-cults/":
        "Una rete di camere antiche sotto le fogne, collegate ai dungeon perduti soprastanti. Passaggi tortuosi conducono verso ovest e verso livelli ancora piu profondi.",

    "https://dysonlogos.blog/2025/08/04/index-card-dungeon-ii-map-10-lost-dungeons/":
        "Sotterranei dimenticati sotto antiche fogne, con una sola scala che conduce ai livelli superiori. Un luogo dove il tempo sembra essersi fermato.",

    "https://dysonlogos.blog/2025/07/28/index-card-dungeon-ii-map-9-ancient-sewers/":
        "Fogne antiche sotto le rovine della torre, ancora percorse da acque scure. I tunnel si estendono in ogni direzione, nascondendo segreti dimenticati.",

    "https://dysonlogos.blog/2025/07/21/index-card-dungeon-ii-map-8-linking-caverns/":
        "Caverne naturali che collegano i livelli superiori a quelli inferiori del complesso. Le pareti umide brillano alla luce delle torce.",

    "https://dysonlogos.blog/2025/07/14/index-card-dungeon-ii-map-7-north-sub-dungeons/":
        "Sotterranei settentrionali sotto la torre in rovina, un dedalo di corridoi e camere che nascondono antichi segreti.",

    "https://dysonlogos.blog/2025/07/07/index-card-dungeon-ii-map-6-dungeon-ruins/":
        "Rovine sotterranee dove cripte e caverne si fondono. Un luogo dove il confine tra costruzione e natura e andato perduto da tempo.",

    "https://dysonlogos.blog/2025/06/30/index-card-dungeon-ii-map-5-tower-dungeons/":
        "I sotterranei della torre, dove cripte antiche si intrecciano con caverne naturali. Passaggi segreti conducono a livelli ancora piu profondi.",

    "https://dysonlogos.blog/2025/06/23/index-card-dungeon-ii-map-4-tower-ruins/":
        "Le rovine della torre si ergono come un monito silenzioso. Le fondamenta ancora reggono, nascondendo accessi ai livelli inferiori.",

    "https://dysonlogos.blog/2025/06/16/index-card-dungeon-ii-map-3-the-ghost-tower/":
        "La Torre Fantasma si staglia contro il cielo, le sue mura ancora intatte nonostante i secoli. Una fortezza che custodisce passaggi verso l'ignoto.",

    "https://dysonlogos.blog/2025/06/09/index-card-dungeon-ii-map-2-northern-ruins/":
        "Rovine settentrionali avvolte dall'edera e dal tempo. Tra le pietre crollate si intravedono ancora gli accessi ai livelli sotterranei.",

    "https://dysonlogos.blog/2025/06/02/index-card-dungeon-ii-map-1-tower-base/":
        "La base della torre, punto di partenza per l'esplorazione del complesso. Fortificazioni ancora solide proteggono l'ingresso ai dungeon sottostanti.",

    "https://dysonlogos.blog/2025/05/07/the-village-of-millbrook-weird-boring-versions/":
        "Il villaggio di Millbrook, un tranquillo insediamento dove la vita scorre lenta. Ma sotto la superficie, qualcosa di strano si agita.",

    "https://dysonlogos.blog/2025/04/11/building-13-knotty-nets-net-making-repairs-and-sales/":
        "Una bottega specializzata nella fabbricazione e riparazione di reti. L'aria sa di canapa e sale marino.",

    "https://dysonlogos.blog/2025/04/07/building-12-general-goods-store/":
        "Un emporio che vende di tutto, dal cibo alle provviste da viaggio. Il proprietario conosce tutti i pettegolezzi del quartiere.",

    "https://dysonlogos.blog/2025/03/10/the-finger/":
        "Una struttura bizzarra e inquietante che si erge come un dito puntato verso il cielo. Al suo interno, scale tortuose conducono verso oscurita indicibili.",

    "https://dysonlogos.blog/2025/02/20/beneath-the-finger/":
        "Sotto Il Dito si estende un complesso sotterraneo dove l'architettura sfida ogni logica. Corridoi impossibili e camere dalle geometrie distorte attendono gli esploratori.",

    "https://dysonlogos.blog/2025/02/13/deeper-beneath-the-finger/":
        "Ancora piu in profondita sotto Il Dito, la realta stessa sembra piegarsi. Le pareti pulsano di un'energia innaturale.",

    "https://dysonlogos.blog/2025/01/28/the-deeps-far-below-the-finger/":
        "Le profondita piu remote sotto Il Dito, dove la follia e la magia si fondono. Pochi sono tornati da questi abissi per raccontare cio che hanno visto.",

    "https://dysonlogos.blog/2025/01/24/a-dungeon-in-two-parts/":
        "Un sotterraneo diviso in due sezioni distinte, collegate da un passaggio stretto. Cripte e tombe si aprono su entrambi i lati.",

    "https://dysonlogos.blog/2024/12/30/eastmeadow-manor/":
        "Un maniero in pietra a tre piani fuori citta, dimora della famiglia Lindwyne e dei loro tre servitori. Sotto la superficie si celano tre livelli di sotterranei.",

    "https://dysonlogos.blog/2024/12/18/under-eastmeadow-manor/":
        "I sotterranei sotto il Maniero di Eastmeadow si estendono su tre livelli. Cosa si nasconde nelle profondita sotto la dimora dei Lindwyne?",

    "https://dysonlogos.blog/2024/12/02/allerton-hold/":
        "Una fortezza in rovina che un tempo dominava la vallata. Le sue mura sgretolate nascondono ancora segreti e passaggi dimenticati.",

    "https://dysonlogos.blog/2025/01/17/crumbling-gate/":
        "Le vecchie rovine tra Redwick Bush e Mournstead sarebbero del tutto insignificanti, se non fosse che il sentiero dei contadini le attraversa. Tre torri, di cui una ancora incombente, creano il luogo perfetto per un'imboscata.",

    "https://dysonlogos.blog/2025/01/14/the-ruined-temple/":
        "Un tempio caduto in rovina, le cui colonne spezzate testimoniano una grandezza passata. Tra le macerie, qualcosa di antico e pericoloso attende.",

    "https://dysonlogos.blog/2024/12/24/beneath-the-black-cairn/":
        "Sotto il tumulo di pietre nere si apre un sepolcro dimenticato. L'oscurita qui ha un peso quasi fisico.",

    "https://dysonlogos.blog/2024/12/11/the-pipeworks/":
        "Un complesso di tubature e condotti che si snoda sotto la superficie, collegando cisterne e camere nascoste. L'acqua scorre ancora in alcuni punti.",

    "https://dysonlogos.blog/2024/11/27/the-sinkhole-crypt/":
        "Una voragine ha rivelato un'antica cripta sepolta. Le pareti franose rendono la discesa pericolosa quanto cio che attende in fondo.",

    "https://dysonlogos.blog/2024/11/18/building-11-carpenter/":
        "La bottega del carpentiere, piena di trucioli e odore di legno fresco. Attrezzi appesi alle pareti e progetti incompiuti ovunque.",

    "https://dysonlogos.blog/2024/11/04/the-greywater-temple/":
        "Un tempio dedicato a divinita dimenticate, le cui fondamenta sono perennemente bagnate da acque grigie e stagnanti. Un'aura di sacralita corrotta permea ogni pietra.",

    "https://dysonlogos.blog/2024/10/30/griffon-tower-dungeons/":
        "I sotterranei della Torre del Grifone si estendono ben oltre le fondamenta. Celle, passaggi segreti e antiche prigioni attendono chi osa scendere.",

    "https://dysonlogos.blog/2024/10/28/griffon-tower/":
        "La Torre del Grifone si erge imponente, un tempo simbolo di potere e ora rifugio di creature oscure. Le sue mura raccontano storie di battaglie e tradimenti.",

    "https://dysonlogos.blog/2024/10/24/ruins-of-the-waste-treatment-facility/":
        "Le rovine di un impianto di trattamento, resti di una civilta piu avanzata. Macchinari arrugginiti e vasche vuote raccontano di un passato perduto.",

    "https://dysonlogos.blog/2024/10/21/building-10-wit-wisdom-apothecary/":
        "Lo speziale Wit & Wisdom, dove erbe rare e pozioni misteriose riempiono scaffali polverosi. Il proprietario sembra sapere sempre esattamente di cosa hai bisogno.",

    "https://dysonlogos.blog/2024/10/17/crypts-and-worms/":
        "Sotto le rovine del tempio si trovano le catacombe dei morti onorati, ormai parzialmente saccheggiate da ladri audaci e infestate da vermi giganteschi.",

    "https://dysonlogos.blog/2024/10/14/the-cinereous-basilica/":
        "Una basilica ricoperta di cenere, i cui affreschi sbiaditi raccontano di una divinita dimenticata. L'aria e densa e grigia, e ogni suono sembra attutito.",

    "https://dysonlogos.blog/2024/10/07/building-9-the-riddle-of-steel/":
        "Una fucina dove il fabbro forgia enigmi nel metallo. Ogni lama racconta una storia, ogni armatura nasconde un segreto.",

    "https://dysonlogos.blog/2024/09/30/the-autumn-lands-map-i/":
        "Le Terre d'Autunno si stendono a sud, un mosaico di colline dorate e foreste dal fogliame infuocato. Una regione da esplorare esagono per esagono.",

    "https://dysonlogos.blog/2024/09/23/the-autumn-lands-map-h/":
        "Un'altra sezione delle Terre d'Autunno, dove sentieri dimenticati serpeggiano tra valli nebbiose e rovine coperte d'edera.",

    "https://dysonlogos.blog/2024/09/16/the-autumn-lands-map-g/":
        "Le Terre d'Autunno continuano a rivelare i loro segreti. Fiumi lenti attraversano pianure dove il vento porta sussurri antichi.",

    "https://dysonlogos.blog/2024/09/09/the-autumn-lands-hex-map-f/":
        "Una mappa esagonale delle Terre d'Autunno, dove ogni esagono nasconde un'avventura. Boschi, paludi e antiche rovine punteggiano il paesaggio.",

    "https://dysonlogos.blog/2024/09/02/the-autumn-lands-hex-map-e/":
        "Le Terre d'Autunno viste dall'alto: colline ondulate, torrenti cristallini e misteriosi cerchi di pietre attendono gli esploratori.",

    "https://dysonlogos.blog/2024/08/26/the-autumn-lands-map-4/":
        "Un tratto delle Terre d'Autunno dove la civilta cede il passo alla natura selvaggia. Pochi sentieri attraversano questa landa desolata.",

    "https://dysonlogos.blog/2024/08/19/the-autumn-lands-map-c/":
        "Le Terre d'Autunno si fanno piu aspre qui, con gole profonde e foreste impenetrabili. Un territorio ideale per imboscate e nascondigli.",

    "https://dysonlogos.blog/2024/08/12/the-autumn-lands-map-b/":
        "Una porzione delle Terre d'Autunno dove antiche strade commerciali incrociano piste di cacciatori. Il paesaggio muta ad ogni passo.",

    "https://dysonlogos.blog/2024/08/05/the-autumn-lands-map-a/":
        "Il primo frammento delle Terre d'Autunno, dove il viaggio ha inizio. Pianure dorate si estendono fino all'orizzonte, interrotte da boschi e villaggi sparsi.",

    "https://dysonlogos.blog/2024/09/25/temple-of-the-worm-upper-levels/":
        "I livelli superiori del Tempio del Verme, dove corridoi ricoperti di muco conducono a sale cerimoniali profanate. L'aria e greve dell'odore della decomposizione.",

    "https://dysonlogos.blog/2024/09/18/temple-of-the-worm/":
        "Un tempio dedicato a una divinita ctonica vermiforme, le cui pareti sono scavate nella roccia viva. I fedeli del Verme compiono i loro rituali nelle profondita.",

    "https://dysonlogos.blog/2024/09/11/building-8-leather-worker/":
        "La bottega del conciatore, dove pelli di ogni tipo pendono ad asciugare. L'odore pungente del tannino pervade l'aria.",

    "https://dysonlogos.blog/2024/08/28/iseldecs-drop-levels-20-23/":
        "I livelli piu profondi della Caduta di Iseldec, dove la roccia stessa sembra respirare. Solo i piu audaci si avventurano a queste profondita.",

    "https://dysonlogos.blog/2024/08/21/brezchyns-halls/":
        "Le Sale di Brezchyn, un complesso di camere e corridoi scolpiti nella roccia con maestria nanica. Nonostante l'abbandono, l'eco di antichi canti risuona ancora.",

    "https://dysonlogos.blog/2024/08/14/building-7-random-market-tavern/":
        "Una taverna nel cuore del mercato, dove mercanti e avventurieri si mescolano tra boccali di birra e voci di tesori nascosti.",

    "https://dysonlogos.blog/2024/08/07/building-6-wine-garden/":
        "Un giardino dei vini dove la vite si arrampica su pergolati antichi. Un luogo tranquillo dove concludere affari o scambiare segreti.",

    "https://dysonlogos.blog/2024/07/31/iseldecs-drop-levels-17-19/":
        "Livelli intermedi della Caduta di Iseldec, dove passaggi naturali e corridoi scavati si intrecciano in un dedalo di pietra e oscurita.",

    "https://dysonlogos.blog/2024/07/24/iseldecs-drop-levels-13-16/":
        "La Caduta di Iseldec si fa piu stretta e insidiosa a queste profondita. Le caverne naturali cedono il passo a costruzioni antiche e dimenticate.",

    "https://dysonlogos.blog/2024/07/17/building-5-loans-bank/":
        "Una banca e agenzia di prestiti, protetta da serrature robuste e guardie discrete. I forzieri nelle cantine contengono piu di semplice oro.",

    "https://dysonlogos.blog/2024/07/10/shop-4-the-fruit-shop/":
        "Un fruttivendolo dai colori vivaci, dove cesti traboccanti di frutta esotica attirano clienti da tutto il quartiere.",

    "https://dysonlogos.blog/2024/07/03/shop-3-carpets/":
        "Una bottega di tappeti dove stoffe preziose ricoprono ogni superficie. Si dice che alcuni tappeti nascondano mappe di tesori perduti.",

    "https://dysonlogos.blog/2024/06/26/iseldecs-drop-levels-9-12/":
        "Livelli della Caduta di Iseldec dove le pareti si restringono e l'aria si fa umida. Stalattiti e stalagmiti formano foreste di pietra.",

    "https://dysonlogos.blog/2024/06/19/stonewill-estate/":
        "La tenuta di Stonewill, un complesso di edifici circondato da giardini curati. Dietro la facciata rispettabile si celano segreti di famiglia.",

    "https://dysonlogos.blog/2024/06/12/the-golden-fishmarket/":
        "Il Mercato del Pesce Dorato, un edificio vivace e rumoroso dove il pescato del giorno viene venduto all'alba. L'odore del mare pervade ogni angolo.",

    "https://dysonlogos.blog/2024/06/05/roks-pottery-workshop/":
        "Il laboratorio di Rok, dove l'argilla prende forme meravigliose sotto le mani del vasaio. Vasi, anfore e statuette riempiono ogni scaffale.",

    "https://dysonlogos.blog/2024/05/29/grudge-match-at-the-underground-market/":
        "Un mercato sotterraneo dove si commercia in merci proibite e si regolano vecchi conti. La tensione nell'aria e palpabile quanto l'odore di spezie esotiche.",

    "https://dysonlogos.blog/2024/05/22/menrinas-library-the-salty-tavern/":
        "La biblioteca di Menrina e la Taverna Salata condividono lo stesso edificio: sapere e spiriti forti sotto un unico tetto.",

    "https://dysonlogos.blog/2024/05/15/dungeon-of-the-bad-egg/":
        "Il Sotterraneo dell'Uovo Marcio, un complesso che deve il suo nome al fetore sulfureo che permea ogni corridoio. Chi vi entra raramente torna con lo stomaco intatto.",

    "https://dysonlogos.blog/2024/05/08/iseldecs-drop-levels-5-8/":
        "La Caduta di Iseldec rivela nuove meraviglie e pericoli a questi livelli. Ponti naturali attraversano voragini e fiumi sotterranei mormorano nell'oscurita.",

    "https://dysonlogos.blog/2024/05/01/ieradess-isle/":
        "L'Isola di Ierades, un frammento di terra emerso dalle acque scure. Rovine antiche si ergono tra la vegetazione lussureggiante.",

    "https://dysonlogos.blog/2024/04/24/goretooths-grotto/":
        "La Grotta di Goretooth, dove stalattiti insanguinate pendono come zanne dal soffitto. Il predatore che da il nome a questo luogo potrebbe ancora aggirarsi nelle ombre.",

    "https://dysonlogos.blog/2024/04/17/sisters-ford/":
        "Un guado lungo il fiume, sorvegliato dalle rovine di due torri gemelle. Un punto di passaggio strategico per viaggiatori e predoni.",

    "https://dysonlogos.blog/2024/04/10/iseldecs-drop-levels-1-4/":
        "I primi livelli della Caduta di Iseldec, dove l'esplorazione ha inizio. Corridoi scavati nella roccia conducono sempre piu in profondita.",

    "https://dysonlogos.blog/2024/04/03/the-tower-of-mourning/":
        "La Torre del Lutto si erge solitaria, avvolta in un silenzio innaturale. Si dice che chi vi entra senta il pianto di coloro che sono stati dimenticati.",

    "https://dysonlogos.blog/2024/03/27/the-temple-of-love-2024-remix/":
        "Un tempio un tempo dedicato all'amore, ora corrotto e profanato. I suoi giardini interni sono infestati da rampicanti velenosi.",

    "https://dysonlogos.blog/2024/03/20/the-sleeping-goat-inn/":
        "La Locanda della Capra Dormiente, un rifugio accogliente per viaggiatori stanchi. La birra e forte, i letti morbidi, e l'oste sa tenere un segreto.",

    "https://dysonlogos.blog/2024/03/13/temple-of-lost-ormus/":
        "Il Tempio del Perduto Ormus giace dimenticato nella foresta, le sue sale un tempo sfarzose ora ricoperte di muschio e radici.",

    "https://dysonlogos.blog/2024/03/06/the-lame-barghest-tavern/":
        "Il Barghest Zoppo e una taverna di dimensioni considerevoli nella Citta Blu, uno dei posti migliori per bere se ci si intende con le bande che garantiscono 'l'ordine pubblico'.",

    "https://dysonlogos.blog/2024/02/28/the-lost-watch/":
        "Una torre di guardia perduta nella nebbia, il cui faro non illumina piu la costa. Qualcosa si muove al suo interno nelle notti senza luna.",

    "https://dysonlogos.blog/2024/02/21/gath-ams-beacon/":
        "Il Faro di Gath-Am, un tempo guida per i naviganti, ora spento e circondato da relitti. La luce che talvolta si intravede non e di questo mondo.",

    "https://dysonlogos.blog/2024/02/14/the-strangled-imp/":
        "Il Diavoletto Strangolato, una taverna dal nome sinistro dove si incontrano personaggi di dubbia reputazione. L'insegna cigola nel vento.",

    "https://dysonlogos.blog/2024/02/07/the-sordid-rhinoceros/":
        "Il Rinoceronte Sordido, un locale malfamato dove l'ale scorre a fiumi e le risse sono l'intrattenimento della serata.",

    "https://dysonlogos.blog/2024/01/31/gnarshs-domain/":
        "Il Dominio di Gnarsh, un complesso di caverne e tunnel reclamato da una creatura di immensa ferocia. Le ossa dei visitatori precedenti decorano l'ingresso.",

    "https://dysonlogos.blog/2024/01/24/the-house-of-kalaxar-linear-dungeon-experiment-2/":
        "La Casa di Kalaxar, un sotterraneo lineare dove ogni stanza conduce inevitabilmente alla successiva. Non c'e via di fuga se non andare avanti.",

    "https://dysonlogos.blog/2024/01/17/pit-of-the-sand-wraiths-linear-dungeon-experiment-1/":
        "La Fossa degli Spettri di Sabbia, un percorso rettilineo che scende nelle viscere della terra. Gli spettri sussurrano promesse di tesori a chi si avventura oltre.",

    "https://dysonlogos.blog/2024/01/10/edge-strip-dungeon/":
        "Un sotterraneo lungo e stretto, scavato lungo il bordo di una scogliera. Un passo falso e la caduta e eterna.",

    "https://dysonlogos.blog/2023/12/27/miserth-keep-main-level/":
        "Il livello principale del Fortino di Miserth, una struttura difensiva che ha visto giorni migliori. Le sue sale un tempo animate ora giacciono in un silenzio carico di tensione.",

    "https://dysonlogos.blog/2023/12/20/miserth-keep-upper-levels/":
        "I livelli superiori del Fortino di Miserth offrono una vista sulla campagna circostante, ma anche un'esposizione pericolosa ai nemici.",

    "https://dysonlogos.blog/2023/12/13/miserth-keep-dungeons/":
        "I sotterranei del Fortino di Miserth, dove prigionieri e segreti venivano custoditi con uguale cura. Le catene arrugginite pendono ancora dalle pareti.",

    "https://dysonlogos.blog/2023/12/06/the-black-skulls-tomb/":
        "La Tomba dei Teschi Neri, un sepolcro profanato dove i resti dei guerrieri caduti riposano accanto a trappole mortali.",

    "https://dysonlogos.blog/2023/11/29/archons-tower/":
        "La Torre dell'Arconte si erge come un monolite di potere arcano. Le sue stanze sono ancora intrise di magia residua.",

    "https://dysonlogos.blog/2023/11/22/beneath-the-archons-tower/":
        "Sotto la Torre dell'Arconte si celano laboratori alchemici e celle di contenimento. Le protezioni magiche si stanno indebolendo.",

    "https://dysonlogos.blog/2023/11/15/desert-clanhold/":
        "Una fortezza scavata nella roccia del deserto, dimora di un clan guerriero. Mura possenti proteggono un'oasi nascosta al suo centro.",

    "https://dysonlogos.blog/2023/11/08/the-deep-sepulchre/":
        "Un sepolcro nelle viscere della terra, dove i morti di un'era dimenticata giacciono in sarcofagi di pietra nera. Il silenzio qui e assoluto.",

    "https://dysonlogos.blog/2023/11/01/temple-of-the-divinity-in-copper/":
        "Un tempio dedicato a una divinita del rame, le cui pareti brillano di una patina verde. I rituali qui praticati promettono ricchezza ma esigono un prezzo.",

    "https://dysonlogos.blog/2023/10/25/pitmann-manse/":
        "La Dimora di Pitmann, una magione decadente dove ogni stanza racconta di ricchezze perdute e segreti di famiglia.",

    "https://dysonlogos.blog/2023/10/18/the-halls-of-lost-heroes/":
        "Le Sale degli Eroi Perduti, dove statue e affreschi celebrano campioni dimenticati dalla storia. I loro spiriti, dicono, non hanno ancora trovato pace.",

    "https://dysonlogos.blog/2023/10/11/bolukbasti-grotto/":
        "La Grotta di Bolukbasti, un rifugio naturale dalle pareti luccicanti di minerali. Le stalattiti creano un labirinto naturale di rara bellezza.",

    "https://dysonlogos.blog/2023/10/04/drurdelm-tombs/":
        "Le Tombe di Drurdelm, un complesso funerario dove generazioni di nobili riposano in camere riccamente decorate. I guardiani non sono tutti di pietra.",

    "https://dysonlogos.blog/2023/09/27/greyrock-tower/":
        "La Torre di Greyrock, una struttura austera di pietra grigia che domina la brughiera. Il suo ultimo occupante non se n'e mai andato.",

    "https://dysonlogos.blog/2023/09/20/statues-of-the-thrall-gods/":
        "Statue colossali di divinita schiave, erette da un popolo che venerava la sottomissione. I loro occhi di pietra sembrano seguire chi passa.",

    "https://dysonlogos.blog/2023/09/13/flooded-catacombs-1200-dpi/":
        "Catacombe parzialmente sommerse, dove l'acqua nera lambisce i sarcofagi e riflette la luce delle torce come uno specchio oscuro.",

    "https://dysonlogos.blog/2023/09/06/sanctuary-of-the-magi-1200-dpi/":
        "Un santuario dove i maghi di un'epoca passata si riunivano per condividere conoscenze arcane. I loro incantesimi proteggono ancora queste sale.",

    "https://dysonlogos.blog/2023/08/30/windsong-hall/":
        "La Sala di Windsong, dove il vento crea melodie inquietanti attraversando corridoi e finestre rotte. Un luogo di bellezza spettrale.",

    "https://dysonlogos.blog/2023/08/23/willowstone-hall/":
        "La Sala di Willowstone, un edificio elegante circondato da salici piangenti. Le sue fondamenta nascondono passaggi verso il sottosuolo.",

    "https://dysonlogos.blog/2023/08/16/the-garden-tower/":
        "La Torre del Giardino, dove la natura ha reclamato ogni livello. Rampicanti e fiori selvatici crescono tra le pietre, creando un giardino verticale.",

    "https://dysonlogos.blog/2023/08/09/gladhold-estate/":
        "La Tenuta di Gladhold, un complesso rurale che nasconde una storia turbolenta. I campi intorno sono fertili, ma la terra sotto la casa e irrequieta.",

    "https://dysonlogos.blog/2023/08/02/the-rumbledown-ruins/":
        "Le Rovine Rumbledown, un ammasso di pietre crollate che un tempo era una dimora signorile. Il terreno trema ancora occasionalmente.",

    "https://dysonlogos.blog/2023/07/27/ency-glowlands-map-2/":
        "La seconda mappa delle Terre Luminose di Ency, dove funghi bioluminescenti illuminano caverne di rara bellezza.",

    "https://dysonlogos.blog/2023/07/26/gascons-pit/":
        "La Fossa di Gascon, una voragine che inghiotte la luce. Chi vi scende trova cunicoli che si ramificano in ogni direzione.",

    "https://dysonlogos.blog/2023/07/19/the-stony-shore-map-1/":
        "La Riva Rocciosa, dove scogliere frastagliate incontrano un mare agitato. Grotte marine si aprono alla base delle rocce.",

    "https://dysonlogos.blog/2023/07/12/the-stony-shore-map-2/":
        "La seconda sezione della Riva Rocciosa, dove sentieri costieri collegano insenature nascoste e rovine marittime.",

    "https://dysonlogos.blog/2023/07/05/the-stony-shore-map-3/":
        "L'ultima sezione della Riva Rocciosa, dove il paesaggio si fa piu selvaggio e le grotte piu profonde.",

    "https://dysonlogos.blog/2023/06/28/the-delren-street-sewers/":
        "Le fogne sotto Via Delren, un labirinto di canali e passaggi che i contrabbandieri conoscono meglio delle guardie cittadine.",

    "https://dysonlogos.blog/2023/06/21/the-whispering-outpost-bw/":
        "L'Avamposto Sussurrante, una fortificazione dove il vento tra le rovine sembra formare parole. I soldati di stanza qui non dormono mai bene.",

    "https://dysonlogos.blog/2023/06/14/bore-facility-p44/":
        "L'Impianto di Trivellazione P44, una struttura tecnologica abbandonata. I macchinari arrugginiti nascondono ancora pericoli funzionanti.",

    "https://dysonlogos.blog/2023/06/07/ziniks-stones/":
        "Le Pietre di Zinik, un circolo megalitico dove le pietre vibrano al tocco. I druidi li evitano, e questo dovrebbe dire qualcosa.",

    "https://dysonlogos.blog/2023/05/31/beneath-the-claw-of-sunsets-map-3/":
        "Le profondita sotto l'Artiglio dei Tramonti, dove tunnel antichi si intrecciano con caverne naturali. La luce del tramonto non arriva mai qui.",

    "https://dysonlogos.blog/2023/05/24/ruins-of-the-claw-of-sunsets-map-3/":
        "Le rovine in superficie dell'Artiglio dei Tramonti, dove mura crollate e torri spezzate raccontano di un'antica fortezza.",

    "https://dysonlogos.blog/2023/05/17/proving-grounds-of-the-mad-ogre-lord/":
        "Il Campo di Prova del Signore Ogre Folle, un'arena brutale dove prigionieri e bestie vengono messi alla prova per il divertimento del loro carceriere.",

    "https://dysonlogos.blog/2023/05/10/the-yellow-fane-in-ruins-1200dpi/":
        "Il Tempio Giallo giace in rovine, le sue mura dorate ormai annerite dal tempo. Tra le macerie, l'altare principale emana ancora un bagliore malsano.",

    "https://dysonlogos.blog/2023/05/03/under-the-lighthouse/":
        "Sotto il faro si nascondono camere e passaggi scavati nella scogliera. Un tempo rifugio per i guardiani, ora dimora di qualcosa di diverso.",

    "https://dysonlogos.blog/2023/04/26/the-dry-lighthouse/":
        "Il Faro Secco, cosi chiamato perche il mare si e ritirato lasciandolo su una pianura arida. La sua luce non serve piu ai naviganti, ma qualcuno continua ad accenderla.",

    "https://dysonlogos.blog/2023/04/19/beneath-the-claw-of-sunsets-map-2/":
        "Il secondo livello sotto l'Artiglio dei Tramonti, dove i tunnel si fanno piu stretti e l'aria piu pesante.",

    "https://dysonlogos.blog/2023/04/12/ruins-of-the-claw-of-sunsets-map-2/":
        "La seconda sezione delle rovine dell'Artiglio dei Tramonti, dove torri diroccate si affacciano su un paesaggio desolato.",

    "https://dysonlogos.blog/2023/04/05/sanctum-cay-bw/":
        "Un isolotto sacro dove un antico santuario resiste all'erosione del mare. Le maree rivelano sentieri nascosti verso grotte sottomarine.",

    "https://dysonlogos.blog/2023/03/29/the-long-halls-of-taqash-thesk/":
        "Le Lunghe Sale di Taqash Thesk si estendono per centinaia di metri sotto la superficie, un complesso monumentale di un'era dimenticata.",

    "https://dysonlogos.blog/2023/03/22/science-tower/":
        "La Torre della Scienza, dove un inventore solitario conduceva esperimenti al limite tra magia e tecnologia. I suoi dispositivi funzionano ancora, in modo imprevedibile.",

    "https://dysonlogos.blog/2023/03/15/opals-steps-stepped-pyramid/":
        "I Gradini di Opal, una piramide a gradoni che si erge dalla giungla. Ogni livello nasconde trappole e camere segrete dedicate a divinita solari.",

    "https://dysonlogos.blog/2023/03/08/beneath-the-claw-of-sunsets-map-1/":
        "Il primo livello sotto l'Artiglio dei Tramonti, dove le fondamenta della fortezza cedono il passo a caverne naturali.",

    "https://dysonlogos.blog/2023/03/01/ruins-of-the-claw-of-sunsets-map-1/":
        "Le prime rovine dell'Artiglio dei Tramonti, una fortezza un tempo imponente ora reclamata dalla natura e dal tempo.",

    "https://dysonlogos.blog/2023/02/22/temple-of-the-communion-of-zeviax/":
        "Il Tempio della Comunione di Zeviax, dove i fedeli si riunivano per rituali di fusione mistica. Le pareti sono ricoperte di simboli che pulsano debolmente.",

    "https://dysonlogos.blog/2023/02/15/dungeoneer-survival-guide-projection-1/":
        "Una proiezione tridimensionale di un complesso sotterraneo, che mostra come i livelli si sovrappongono e si intrecciano.",

    "https://dysonlogos.blog/2023/02/08/dungeons-of-the-twin-demons/":
        "I Sotterranei dei Demoni Gemelli, un complesso speculare dove ogni ala riflette l'altra. Due entita malvagie si contendono il dominio di queste sale.",

    "https://dysonlogos.blog/2023/02/01/gateway-of-the-twin-demons/":
        "Il Portale dei Demoni Gemelli, un ingresso monumentale fiancheggiato da statue demoniache. Oltre la soglia, il calore aumenta e l'aria vibra.",

    "https://dysonlogos.blog/2023/01/25/the-sirens-grotto/":
        "La Grotta delle Sirene, dove il canto delle onde si mescola a voci soprannaturali. I marinai sanno di evitare questa insenatura.",

    "https://dysonlogos.blog/2023/01/18/the-maw/":
        "Le Fauci, un'apertura nella roccia che sembra la bocca di una creatura colossale. Chiunque entri ha la sensazione di essere inghiottito.",

    "https://dysonlogos.blog/2023/01/11/pits-of-the-black-moon-bw/":
        "Le Fosse della Luna Nera, voragini circolari che si aprono sotto una luna perennemente oscurata. Il fondo non si vede, ma qualcosa si muove la sotto.",

    "https://dysonlogos.blog/2023/01/04/lost-tomb-complex/":
        "Un complesso tombale perduto nella giungla, dove sacerdoti-re riposano circondati dai loro tesori. Le trappole sono ancora attive.",

    "https://dysonlogos.blog/2022/12/28/the-ioun-tower/":
        "La Torre di Ioun, dedicata alla divinita della conoscenza. Ogni piano custodisce tomi e artefatti di sapere proibito.",

    "https://dysonlogos.blog/2022/12/21/hnallas-tower-of-pillars/":
        "La Torre dei Pilastri di Hnálla, un santuario verticale dove colonne di pietra scolpita si innalzano per decine di metri.",

    "https://dysonlogos.blog/2022/12/14/the-grotto-beneath-lazuni-hill/":
        "Una grotta nascosta sotto la Collina Lazuni, accessibile solo a chi conosce il sentiero segreto tra le rocce.",

    "https://dysonlogos.blog/2022/12/07/the-shrine-atop-lazuni-hill/":
        "Il Santuario sulla cima della Collina Lazuni, un luogo sacro esposto ai venti dove i pellegrini vengono a pregare.",

    "https://dysonlogos.blog/2022/11/30/epharas-alithinero-1200-dpi/":
        "L'Alithinero di Ephara, un tempio dedicato alla dea della civilta. Le sue colonne marmoree sfidano il tempo.",

    "https://dysonlogos.blog/2022/11/23/violet-chambers-atop-kakri-midallu-1200-dpi/":
        "Le Camere Viola in cima a Kákri Midállu, sale cerimoniali dove la luce filtra attraverso cristalli ametista.",

    "https://dysonlogos.blog/2022/11/16/the-eleint-passages-ruined-fane/":
        "Un tempio in rovina lungo i Passaggi di Eleint, dove il culto di una divinita dimenticata attendeva il proprio destino.",

    "https://dysonlogos.blog/2022/11/09/the-eleint-passages-the-tower/":
        "Una torre lungo i Passaggi di Eleint, punto di osservazione e difesa lungo questa antica via sotterranea.",

    "https://dysonlogos.blog/2022/11/02/the-eleint-passages-south/":
        "La sezione meridionale dei Passaggi di Eleint, dove i tunnel si allargano in caverne naturali.",

    "https://dysonlogos.blog/2022/10/26/the-eleint-passages-north/":
        "La sezione settentrionale dei Passaggi di Eleint, dove antiche incisioni decorano le pareti di pietra levigata.",

    "https://dysonlogos.blog/2022/10/19/the-basilisks-caves/":
        "Le Caverne del Basilisco, un sistema di grotte dove lo sguardo pietrificante del mostro ha lasciato il segno: statue di pietra di avventurieri sfortunati punteggiano ogni sala.",

    "https://dysonlogos.blog/2022/10/12/cult-basement/":
        "Il seminterrato segreto di un culto, nascosto sotto un edificio apparentemente innocuo. Simboli arcani ricoprono le pareti e l'altare.",

    "https://dysonlogos.blog/2022/10/05/chambers-of-the-copper-skulls/":
        "Le Camere dei Teschi di Rame, sale sotterranee decorate con teschi ricoperti di rame che brillano alla luce delle torce.",

    "https://dysonlogos.blog/2022/09/28/ruins-of-the-hill-fort/":
        "Le rovine di un forte collinare, le cui mura offrivano un tempo una vista strategica sulla valle. Ora solo i corvi fanno la guardia.",

    "https://dysonlogos.blog/2022/09/21/temple-cave-of-the-ruinous-ministers/":
        "Una grotta sacra dove ministri eretici celebravano rituali di distruzione. Le pareti sono annerite da fuochi rituali.",

    "https://dysonlogos.blog/2022/09/14/vund-home/":
        "La Dimora di Vund, una casa modesta che cela passaggi segreti e stanze nascoste. Il precedente proprietario non era chi diceva di essere.",

    "https://dysonlogos.blog/2022/09/07/the-ency-glowlands-map-1/":
        "La prima mappa delle Terre Luminose di Ency, un ecosistema sotterraneo illuminato da funghi bioluminescenti e cristalli radianti.",

    "https://dysonlogos.blog/2022/08/31/dungeon-of-holding/":
        "Il Sotterraneo della Custodia, un complesso progettato per contenere qualcosa di terribile. Le celle sono intatte, ma i loro occupanti?",

    "https://dysonlogos.blog/2022/08/24/sharn-skybridges-iii/":
        "Ponti celesti che collegano le torri vertiginose di Sharn, sospesi centinaia di metri sopra il suolo. Un passo falso e fatale.",

    "https://dysonlogos.blog/2022/08/17/dead-goblin-hole/":
        "La Tana del Goblin Morto, cosi chiamata per ovvie ragioni. Ma la domanda e: cosa ha ucciso il goblin, e si trova ancora li dentro?",

    "https://dysonlogos.blog/2022/08/10/bradrigs-hall/":
        "La Sala di Bradrig, un tempo luogo di banchetti e celebrazioni. Ora le lunghe tavole sono rovesciate e le pareti macchiate.",

    "https://dysonlogos.blog/2022/08/03/ruined-cities-of-yoth/":
        "Le Citta in Rovina di Yoth, vestigia di una civilta serpentina che un tempo dominava il mondo sotterraneo.",

    "https://dysonlogos.blog/2022/07/27/smallharbour/":
        "Porto Piccolo, un insediamento costiero dove pescatori e contrabbandieri convivono in precario equilibrio.",

    "https://dysonlogos.blog/2022/07/20/hunters-descent/":
        "La Discesa del Cacciatore, un passaggio ripido che conduce nelle profondita della terra. Il cacciatore e diventato la preda.",

    "https://dysonlogos.blog/2022/07/13/the-idol-pit/":
        "La Fossa dell'Idolo, dove una statua di una divinita sconosciuta giace in fondo a un cratere. I cultisti la venerano ancora.",

    "https://dysonlogos.blog/2022/07/06/sanctuary-of-xeeus/":
        "Il Santuario di Xeeus, un luogo di meditazione arcana nascosto nelle profondita. Le sue protezioni magiche sono ancora attive.",

    "https://dysonlogos.blog/2022/06/29/beneath-the-temperance-stone/":
        "Sotto la Pietra della Temperanza si apre un dungeon a cinque stanze, con un menhir a guardia dell'ingresso come in molti complessi simili.",

    "https://dysonlogos.blog/2022/06/22/tombs-of-the-silver-army/":
        "Le Tombe dell'Esercito d'Argento, dove guerrieri sepolti in armature argentate riposano in formazione di battaglia. Si dice che al suono di un corno, si risveglieranno.",

    "https://dysonlogos.blog/2022/06/15/garnets-spring/":
        "La Sorgente di Garnet, dove acque cristalline emergono dalla roccia tinta di rosso. Un luogo pacifico, se non fosse per cio che vive nelle profondita.",

    "https://dysonlogos.blog/2022/06/08/a-dungeon-of-impossible-stairs-1200-dpi/":
        "Un sotterraneo dove le scale sfidano la gravita e la geometria. Salire puo significare scendere, e le porte non conducono dove ci si aspetta.",

    "https://dysonlogos.blog/2022/06/01/gavreths-stones/":
        "Le Pietre di Gavreth, un antico circolo cerimoniale dove il velo tra i mondi e sottile. Durante le notti di luna piena, le pietre cantano.",

    "https://dysonlogos.blog/2022/05/25/veghuls-drop/":
        "Il Precipizio di Veghul, una cascata verticale che conduce a caverne inesplorate. L'eco dell'acqua maschera altri suoni piu inquietanti.",

    "https://dysonlogos.blog/2022/05/18/hardbrand-tower/":
        "La Torre di Hardbrand, costruita da un cavaliere paranoico con trappole ad ogni piano. Le sue ricchezze sono ancora al suo interno.",

    "https://dysonlogos.blog/2022/05/11/the-last-stones-of-sagemane-castle/":
        "Le ultime pietre del Castello di Sagemane emergono dalla vegetazione come denti rotti. Solo le fondamenta e le cantine rimangono intatte.",

    "https://dysonlogos.blog/2022/05/04/the-jade-catacombs/":
        "Le Catacombe di Giada, dove i sarcofagi sono scolpiti in nefrite verde. Una luce smeraldina pervade queste sale silenziose.",

    "https://dysonlogos.blog/2022/04/27/the-bloodied-stones-a-druidic-circle/":
        "Un circolo druidico dove le pietre sono macchiate di un rosso che non sbiadisce mai. I rituali qui praticati legano la terra al sangue.",

    "https://dysonlogos.blog/2022/04/20/tombs-of-the-throl-tribe/":
        "Le Tombe della Tribu Throl, sepolcri di guerrieri barbari sepolti con le loro armi. I tumuli sono disposti secondo costellazioni dimenticate.",

    "https://dysonlogos.blog/2022/04/13/the-andesite-pyramid/":
        "La Piramide di Andesite, una struttura massiccia di pietra vulcanica grigia. I suoi corridoi interni seguono un percorso a spirale verso il cuore.",

    "https://dysonlogos.blog/2022/04/06/the-red-temple-under-bey-su/":
        "Il Tempio Rosso sotto Béy Sǘ, un santuario sotterraneo le cui pareti sono dipinte di cremisi. I sacerdoti che lo abitavano servivano un dio della guerra.",

    "https://dysonlogos.blog/2022/03/30/barrow-of-the-flayer/":
        "Il Tumulo dello Scorticatore, un sepolcro maledetto che nessuno del villaggio vicino osa avvicinare. Si sentono grida nelle notti senza vento.",

    "https://dysonlogos.blog/2022/03/23/release-the-kraken-on-the-warkings-vault/":
        "La Volta dei Warkings, un deposito blindato dove generazioni di mercanti hanno accumulato ricchezze e segreti.",

    "https://dysonlogos.blog/2022/03/16/release-the-kraken-on-the-mastervale-estate/":
        "La Tenuta di Mastervale, un complesso signorile con giardini formali e cantine labirintiche. La famiglia che la abitava e scomparsa misteriosamente.",

    "https://dysonlogos.blog/2022/03/09/the-village-of-winten-ford/":
        "Il Villaggio di Winten Ford, un insediamento tranquillo presso un guado del fiume. La vita scorre placida, almeno in superficie.",

    "https://dysonlogos.blog/2022/03/02/the-maze-of-yivhkthaloth/":
        "Il Labirinto di Yivh'Kthaloth, un dedalo vivente che muta forma secondo la volonta di un'entita aliena. Chi vi entra non trova mai la stessa via due volte.",

    "https://dysonlogos.blog/2022/02/23/dohrlegonds-tombs/":
        "Le Tombe di Dohrlegond, camere funerarie di un antico signore della guerra. I suoi generali giacciono con lui, pronti a servirlo anche nella morte.",

    "https://dysonlogos.blog/2022/02/16/passages-beneath/":
        "I Passaggi Sottostanti, una rete di tunnel che collega cantine e cripte sotto la citta. Chi li conosce puo attraversare l'intero quartiere senza mai vedere la luce del sole.",

    "https://dysonlogos.blog/2022/02/09/screaming-hall-of-the-ur-goblin/":
        "La Sala Urlante dell'Ur-Goblin, dove l'eco amplifica ogni suono in un grido insopportabile. Il mostro ancestrale che l'ha costruita ama i visitatori rumorosi.",

    "https://dysonlogos.blog/2022/02/02/gaerborin-keep/":
        "Il Fortino di Gaerborin, una struttura difensiva in posizione strategica. Le sue mura hanno respinto assalti per generazioni.",

    "https://dysonlogos.blog/2022/01/26/raining-cave/":
        "La Caverna della Pioggia, dove gocce d'acqua cadono incessantemente dal soffitto creando pozze cristalline. La congregazione eretica del profeta Nezek si e rifugiata qui.",

    "https://dysonlogos.blog/2022/01/19/blackstone-bastion/":
        "Il Bastione di Pietranera, una fortezza costruita con blocchi di ossidiana che assorbono la luce. Un luogo di potere oscuro.",

    "https://dysonlogos.blog/2022/01/12/oreneys-watch/":
        "La Vedetta di Oreney, un posto di guardia isolato su una cresta rocciosa. L'ultima linea di difesa prima delle terre selvagge.",

    "https://dysonlogos.blog/2022/01/05/crypt-of-the-queen-of-bones/":
        "La Cripta della Regina delle Ossa, un mausoleo grandioso dove una sovrana non-morta regna sui resti dei suoi sudditi.",

    "https://dysonlogos.blog/2021/12/29/sumerian-three-story-home/":
        "Una dimora sumera a tre piani, ricostruzione fedele di un'abitazione antica. Le stanze strette e i muri spessi raccontano di un'epoca lontana.",

    "https://dysonlogos.blog/2021/12/22/frogsmead-inn-tavern/":
        "La Locanda e Taverna Frogsmead, un edificio accogliente dove birra forte e storie incredibili scorrono in ugual misura.",

    "https://dysonlogos.blog/2021/12/15/page-o-little-ruins/":
        "Una raccolta di piccole rovine, frammenti di civilta passate sparsi per la campagna. Ognuna racconta una storia diversa.",

    "https://dysonlogos.blog/2021/12/08/under-the-observatory/":
        "Sotto l'osservatorio si nasconde un complesso di stanze dedicate a studi proibiti. Le stelle osservate da qui non sono quelle del nostro cielo.",

    "https://dysonlogos.blog/2021/12/01/unrol-kazad-watch/":
        "La Vedetta di Unrol Kaz'ad, un avamposto nanico intagliato nella montagna. Le sue sale di guardia non sono mai state violate.",

    "https://dysonlogos.blog/2021/11/24/temple-ruins-froghemoth-nest/":
        "Le rovine di un tempio dove un froghemoth ha costruito il proprio nido tra le colonne spezzate. I resti delle offerte si mescolano a quelli delle prede.",

    "https://dysonlogos.blog/2021/11/17/the-sunken-tower/":
        "La Torre Sommersa, sprofondata nel terreno paludoso fino al secondo piano. Solo i piani superiori emergono, mentre sotto il fango si celano i suoi segreti.",

    "https://dysonlogos.blog/2021/11/10/the-goblin-warrens-at-fort-redshield/":
        "Le tane dei goblin scavate sotto il Forte Redshield, un dedalo di tunnel stretti e camere puzzolenti. I goblin hanno trasformato le vecchie cantine in una piccola citta.",

    "https://dysonlogos.blog/2021/11/03/ruins-of-brollmoreth/":
        "Le Rovine di Brollmoreth, resti di una citta che la foresta ha reclamato. Alberi crescono attraverso pavimenti di pietra e radici avvolgono le colonne.",

    "https://dysonlogos.blog/2021/10/27/ruins-of-the-east-gate-of-steldin-dorg/":
        "Le Rovine della Porta Orientale di Steldin Dorg, un tempo ingresso monumentale della citta nanica. Le incisioni runiche sono ancora leggibili.",

    "https://dysonlogos.blog/2021/10/20/the-ruins-of-shagunat-keep/":
        "Le Rovine del Fortino di Shagunat, dove le mura crollate creano un labirinto naturale. Qualcosa si aggira tra le pietre.",

    "https://dysonlogos.blog/2021/10/13/geomorphic-halls-level-1/":
        "Sale geomorfiche che cambiano configurazione ad ogni visita. Nessuna mappa e affidabile in questo luogo dove i corridoi si spostano.",

    "https://dysonlogos.blog/2021/10/06/adventures-around-jalovhec-bw/":
        "Una mappa regionale delle terre intorno a Jalovhec, dove avventure attendono dietro ogni collina e in ogni foresta.",

    "https://dysonlogos.blog/2021/09/29/redford-citadel/":
        "La Cittadella di Redford, una fortezza imponente che controlla il passaggio tra le montagne. Le sue torri rosse sono visibili da leghe di distanza.",

    "https://dysonlogos.blog/2021/09/22/nuzur-hollow/":
        "La Conca di Nuzur, una depressione naturale nel terreno dove la nebbia non si alza mai del tutto. Le rovine al centro sono piu antiche di qualsiasi ricordo.",

    "https://dysonlogos.blog/2021/09/15/darklingtown-the-upper-tunnels/":
        "I tunnel superiori di Darklingtown, dove la luce del mondo esterno filtra ancora debolmente. Il confine tra superficie e sottosuolo.",

    "https://dysonlogos.blog/2021/09/08/darklingtown-cavern-spillways-district/":
        "Il Distretto delle Caverne e degli Sfioratori di Darklingtown, dove l'acqua scorre attraverso canali artificiali tra stalattiti e costruzioni.",

    "https://dysonlogos.blog/2021/09/01/the-dwarven-shrine-at-mount-thorrien/":
        "Il Santuario Nanico al Monte Thorrien, un luogo sacro scolpito nel cuore della montagna. Le rune brillano di luce propria.",

    "https://dysonlogos.blog/2021/08/25/ruldroc-castle/":
        "Il Castello di Ruldroc, una fortezza massiccia con mura spesse tre metri. Le sue segrete non hanno mai rilasciato un prigioniero.",

    "https://dysonlogos.blog/2021/08/18/oceanwatch/":
        "La Vedetta sull'Oceano, un avamposto affacciato sulle onde infinite. Da qui si scrutano i mari per navi nemiche e creature marine.",

    "https://dysonlogos.blog/2021/08/11/darklingtown-the-tunnels-district/":
        "Il Distretto dei Tunnel di Darklingtown, un intrico di passaggi dove la comunita sotterranea vive e commercia.",

    "https://dysonlogos.blog/2021/08/04/darklingtown-frog-tower/":
        "La Torre della Rana di Darklingtown, una struttura bizzarra decorata con motivi anfibi. I suoi abitanti hanno gusti peculiari.",

    "https://dysonlogos.blog/2021/07/28/darklingtown-the-mushroom-caves/":
        "Le Caverne dei Funghi di Darklingtown, dove funghi giganti illuminano l'oscurita con la loro bioluminescenza. Un giardino sotterraneo di bellezza aliena.",

    "https://dysonlogos.blog/2021/07/21/sharn-heights-skybridge-nexus/":
        "Il Nodo dei Ponti Celesti di Sharn, dove molteplici ponti si incrociano a centinaia di metri d'altezza. Un crocevia vertiginoso.",

    "https://dysonlogos.blog/2021/07/14/krelava-manor/":
        "Il Maniero di Krelava, una dimora signorile che nasconde laboratori alchemici nelle cantine. Il padrone di casa e un uomo dai molti segreti.",

    "https://dysonlogos.blog/2021/07/07/daglans-cave/":
        "La Caverna di Daglan, una grotta che serve da rifugio e nascondiglio. Piccola ma ricca di angoli nascosti.",

    "https://dysonlogos.blog/2021/06/30/sewer-complex/":
        "Un complesso fognario intricato sotto la citta, dove i tunnel di scarico si intrecciano con passaggi piu antichi.",

    "https://dysonlogos.blog/2021/06/23/darklingtown-east-frogsport/":
        "La parte orientale di Frogsport, quartiere di Darklingtown dove il commercio fiorisce nonostante l'oscurita perenne.",

    "https://dysonlogos.blog/2021/06/16/darklingtown-frogsport/":
        "Frogsport, il cuore commerciale di Darklingtown. Bancarelle e botteghe illuminate da funghi luminosi fiancheggiano le vie principali.",

    "https://dysonlogos.blog/2021/06/09/bastion-of-the-prince-of-clubs/":
        "Il Bastione del Principe di Bastoni, una fortezza dove il gioco d'azzardo e legge e la sorte decide la vita e la morte.",

    "https://dysonlogos.blog/2021/06/02/kamrorths-cairn/":
        "Il Tumulo di Kamrorth, un sepolcro antico dove un eroe dimenticato riposa con il suo tesoro. I non-morti montano la guardia.",

    "https://dysonlogos.blog/2021/05/26/pirate-booty-island/":
        "L'Isola del Tesoro dei Pirati, dove generazioni di pirati hanno sepolto i loro bottini. X segna il punto, ma quale X?",

    "https://dysonlogos.blog/2021/05/19/darklingtown-south/":
        "Il quartiere meridionale di Darklingtown, dove i tunnel piu antichi si perdono nell'oscurita. Pochi si avventurano oltre le ultime lanterne.",

    "https://dysonlogos.blog/2021/05/12/darklingtown-north-district/":
        "Il Distretto Nord di Darklingtown, il piu vicino alla superficie. Qui l'aria e quasi fresca e la luce quasi naturale.",

    "https://dysonlogos.blog/2021/05/05/imp-tower/":
        "La Torre del Diavoletto, una costruzione snella e tortuosa che sembra sfidare la gravita. I suoi piccoli occupanti sono dispettosi e pericolosi.",

    "https://dysonlogos.blog/2021/04/28/creeping-sands/":
        "Le Sabbie Striscianti, un'area desertica dove le dune si muovono come creature vive, rivelando e nascondendo rovine sepolte.",

    "https://dysonlogos.blog/2021/04/21/business-cards-2021-set-1/":
        "Geomorfi su formato biglietto da visita, moduli di dungeon componibili che si incastrano per creare complessi sotterranei unici.",

    "https://dysonlogos.blog/2021/04/14/2021-business-card-geomorphs-set-2-promotional/":
        "Il secondo set di geomorfi su biglietto, con nuove configurazioni di corridoi e stanze per espandere i sotterranei.",

    "https://dysonlogos.blog/2021/04/07/the-goblin-vault/":
        "La Volta dei Goblin, dove le piccole creature hanno accumulato tutto cio che hanno rubato. Un tesoro caotico e ben difeso.",

    "https://dysonlogos.blog/2021/03/31/heart-of-darkling-the-stairs/":
        "Le Scale del Cuore di Darkling, un sistema di gradini che collega i livelli superiori alle profondita. Ogni pianerottolo nasconde un pericolo.",

    "https://dysonlogos.blog/2021/03/24/the-champions-retreat/":
        "Il Rifugio del Campione, un luogo appartato dove un eroe leggendario trascorse i suoi ultimi giorni. Le sue armi riposano ancora qui.",

    "https://dysonlogos.blog/2021/03/17/mayers-fort/":
        "Il Forte di Mayer, un avamposto militare costruito in fretta ma progettato con astuzia. Le sue difese compensano le dimensioni ridotte.",

    "https://dysonlogos.blog/2021/03/10/nephilims-hall/":
        "La Sala del Nephilim, costruita per esseri di statura superiore. I soffitti altissimi e le porte gigantesche ricordano chi l'ha edificata.",

    "https://dysonlogos.blog/2021/03/03/skaldons-dome/":
        "La Cupola di Skaldon, una struttura emisferica perfetta che sfida ogni spiegazione architettonica. Al suo interno, l'acustica amplifica ogni sussurro.",

    "https://dysonlogos.blog/2021/02/24/pentagonal-monuments-in-the-ghost-dunes/":
        "Monumenti pentagonali emergono dalle Dune Fantasma, strutture geometriche perfette in un mare di sabbia che si muove da sola.",

    "https://dysonlogos.blog/2021/02/17/heart-of-darkling-the-pillars/":
        "I Pilastri del Cuore di Darkling, colonne naturali di pietra che sorreggono una volta immensa. Tra di essi si annidano creature dell'oscurita.",

    "https://dysonlogos.blog/2021/02/10/greyfalls-cave/":
        "La Caverna di Greyfalls, dove una cascata grigia precipita nell'oscurita. Il rombo dell'acqua copre ogni altro suono.",

    "https://dysonlogos.blog/2021/02/03/catacombs-of-the-flayed-minotaur/":
        "Le Catacombe del Minotauro Scorticato, un labirinto funerario dove i resti del mostro sono venerati come reliquie sacre.",

    "https://dysonlogos.blog/2021/01/27/the-dragon-shrine-1200-dpi/":
        "Il Santuario del Drago, un tempio dedicato al culto dei draghi. Le pareti sono decorate con scaglie d'oro e gli occhi delle statue brillano di rubino.",

    "https://dysonlogos.blog/2021/01/20/the-tyrants-ruins/":
        "Le Rovine del Tiranno, resti del palazzo di un despota caduto. Le sue sale del trono sono ancora macchiate dal sangue dell'ultima rivolta.",

    "https://dysonlogos.blog/2021/01/13/the-ritual-pool/":
        "La Pozza del Rituale, uno specchio d'acqua perfettamente circolare in una caverna sotterranea. Le sue acque mostrano visioni a chi osa guardare.",

    "https://dysonlogos.blog/2021/01/06/heart-of-darkling-deceptions-bridge/":
        "Il Ponte dell'Inganno nel Cuore di Darkling, una struttura che sembra solida ma nasconde trabocchetti per gli incauti.",

    "https://dysonlogos.blog/2020/12/30/collapsed-tomb-of-mosogret-1200-dpi/":
        "La Tomba Crollata di Mosogret, dove un terremoto ha rivelato camere sepolcrali sigillate da millenni. I tesori sono intatti, ma anche i guardiani.",

    "https://dysonlogos.blog/2020/12/23/lords-of-the-aldeiron-peaks-1200-dpi/":
        "I Signori delle Vette di Aldeiron, un complesso montano dove antichi signori della guerra riposano in sale scavate nella roccia.",

    "https://dysonlogos.blog/2020/12/16/undercrofts-1200-dpi/":
        "Le Cripte Sotterranee, un livello nascosto sotto una chiesa dove segreti e reliquie sono custoditi lontano da occhi indiscreti.",

    "https://dysonlogos.blog/2020/12/09/the-fallen-house-of-githaleon-1200-dpi/":
        "La Casa Caduta di Githaleon, una dimora elfica crollata in una voragine. Le sue sale eleganti ora giacciono inclinate e rotte.",

    "https://dysonlogos.blog/2020/12/02/heart-of-darkling-the-cold-caves/":
        "Le Caverne Fredde del Cuore di Darkling, dove il ghiaccio ricopre ogni superficie e il respiro si congela all'istante.",

    "https://dysonlogos.blog/2020/11/25/heart-of-darkling-the-drop/":
        "Il Precipizio del Cuore di Darkling, una caduta verticale nell'oscurita piu profonda. Il fondo e cosi lontano che il suono non risale.",

    "https://dysonlogos.blog/2020/11/18/red-talons-lair/":
        "La Tana dell'Artiglio Rosso, rifugio di un predatore leggendario. Le pareti sono segnate da graffi profondi e il terreno cosparso di ossa.",

    "https://dysonlogos.blog/2020/11/11/the-granite-shore-1200-dpi/":
        "La Riva di Granito, una costa rocciosa battuta dalle onde dove grotte marine nascondono tesori di navi naufragate.",

    "https://dysonlogos.blog/2020/11/04/the-dragon-temple-1200-dpi/":
        "Il Tempio del Drago, un santuario monumentale dove un culto draconico compie i propri rituali tra colonne scolpite a forma di serpenti alati.",

    "https://dysonlogos.blog/2020/10/28/the-four-toothed-drunk-1200-dpi/":
        "L'Ubriaco dai Quattro Denti, una taverna di pessima reputazione ma ottima birra. Le risse qui sono quasi un rituale serale.",

    "https://dysonlogos.blog/2020/10/21/heart-of-darkling-gibberling-lake-1200-dpi/":
        "Il Lago dei Gibberling nel Cuore di Darkling, un corpo d'acqua sotterraneo le cui rive pullulano di creature deliranti.",

    "https://dysonlogos.blog/2020/10/14/heart-of-darkling-the-darkling-galleries/":
        "Le Gallerie Darkling, vasti spazi aperti nel cuore delle profondita dove formazioni rocciose creano un paesaggio alieno.",

    "https://dysonlogos.blog/2020/10/07/lost-city-of-the-naga-queens/":
        "La Citta Perduta delle Regine Naga, un impero sotterraneo dove le sovrane serpentine regnano ancora sui loro sudditi immortali.",

    "https://dysonlogos.blog/2020/09/30/island-tomb/":
        "Una tomba solitaria su un'isola remota, raggiungibile solo con la bassa marea. Il defunto voleva essere sicuro di non ricevere visite.",

    "https://dysonlogos.blog/2020/09/23/the-bottomless-tombs-part-4-1200-dpi/":
        "Il quarto livello delle Tombe Senza Fondo, dove la profondita diventa vertigine e le pareti sembrano chiudersi su chi scende.",

    "https://dysonlogos.blog/2020/09/16/lost-reliquary-1200-dpi/":
        "Il Reliquiario Perduto, una camera blindata dove reliquie sacre erano custodite lontano da mani indegne. Le protezioni sono ancora attive.",

    "https://dysonlogos.blog/2020/09/09/landing-facility-with-grids/":
        "Una struttura di atterraggio di origine sconosciuta, con piattaforme e hangar progettati per veicoli che non esistono piu.",

    "https://dysonlogos.blog/2020/09/02/temple-of-the-4-gods-ground-level/":
        "Il Tempio dei 4 Dei, dove quattro divinita rivali condividono un unico luogo sacro. Ogni ala riflette il carattere del suo dio.",

    "https://dysonlogos.blog/2020/08/26/tarsakh-manor-upper-floors-1200-dpi/":
        "I piani superiori del Maniero di Tarsakh, con camere da letto lussuose e una biblioteca fornita. Ma di notte, passi risuonano dove nessuno cammina.",

    "https://dysonlogos.blog/2020/08/19/tarsakh-manor-grounds-1200-dpi/":
        "I terreni del Maniero di Tarsakh, con giardini curati, stalle e un pozzo che dicono sia senza fondo.",

    "https://dysonlogos.blog/2020/08/12/tarsakh-village-1200-dpi/":
        "Il Villaggio di Tarsakh, una comunita agricola all'ombra del maniero. Gli abitanti sono ospitali ma evasivi riguardo al loro signore.",

    "https://dysonlogos.blog/2020/08/05/forlorn-halls-of-the-mongrelfolk-1200-dpi/":
        "Le Sale Desolate dei Mongrelfolk, corridoi abbandonati dove creature ibride un tempo trovavano rifugio dalla persecuzione del mondo esterno.",

    "https://dysonlogos.blog/2020/07/29/twilight-descent-1200-dpi/":
        "La Discesa del Crepuscolo, un passaggio che scende nell'oscurita dove la luce del giorno non raggiunge mai. Ogni passo porta piu vicino all'ignoto.",

    "https://dysonlogos.blog/2020/07/22/return-to-durahns-tomb/":
        "Ritorno alla Tomba di Durahn, dove le trappole sono state resettate e nuovi occupanti si sono insediati tra i sarcofagi.",

    "https://dysonlogos.blog/2020/07/15/workshop-and-store/":
        "Un laboratorio e negozio dove artigianato e commercio si incontrano. Attrezzi, materiali e prodotti finiti riempiono ogni angolo.",

    "https://dysonlogos.blog/2020/07/08/tunnels-of-the-shrouded-emperor-1200-dpi/":
        "I Tunnel dell'Imperatore Velato, passaggi segreti dove un sovrano nascosto muoveva i fili del potere lontano da occhi indiscreti.",

    "https://dysonlogos.blog/2020/07/01/well-gurath-1200-dpi/":
        "Il Pozzo di Gurath, una discesa verticale nel cuore della montagna dove l'acqua scorre lungo pareti di ossidiana.",

    "https://dysonlogos.blog/2020/06/24/the-old-throne-1200-dpi/":
        "Il Vecchio Trono, una sala del trono abbandonata dove il sedile del potere si erge ancora, coperto di ragnatele e polvere secolare.",

    "https://dysonlogos.blog/2020/06/17/skull-maze-1200-dpi/":
        "Il Labirinto del Teschio, un dedalo di corridoi la cui pianta, vista dall'alto, forma un teschio ghignante.",

    "https://dysonlogos.blog/2020/06/10/forbidden-halls/":
        "Le Sale Proibite, camere sigillate da incantesimi che promettono morte a chi le viola. Cosa custodiscono di cosi prezioso?",

    "https://dysonlogos.blog/2020/06/03/alturiak-manor-1200-dpi/":
        "Il Maniero di Alturiak, una dimora gelida come il mese invernale da cui prende il nome. Le correnti d'aria nascondono voci.",

    "https://dysonlogos.blog/2020/05/27/dungeons-of-the-iron-star-level-2-1200-dpi/":
        "Il secondo livello dei Sotterranei della Stella di Ferro, dove le forze infernali hanno esteso la loro presa corrompendo ogni pietra.",

    "https://dysonlogos.blog/2020/05/20/borderlands-caves-level-2-east-1200-dpi/":
        "Il secondo livello delle Caverne delle Terre di Confine, sezione est. Tunnel naturali si intrecciano con scavi antichi.",

    "https://dysonlogos.blog/2020/05/13/quellport-the-isle-of-seven-bees-1200-dpi/":
        "Quellport e l'Isola delle Sette Api, un porto e la sua isola misteriosa dove sette alveari giganti producono un miele dalle proprieta magiche.",

    "https://dysonlogos.blog/2020/05/06/sietch-of-morning-1200-dpi/":
        "Il Sietch del Mattino, un rifugio desertico nascosto tra le rocce dove una comunita vive secondo antiche tradizioni.",

    "https://dysonlogos.blog/2020/04/29/abbey-of-the-iron-star-1200-dpi/":
        "L'Abbazia della Stella di Ferro, un tempo luogo di preghiera ora corrotto da forze infernali. I monaci caduti servono nuovi padroni.",

    "https://dysonlogos.blog/2020/04/22/dungeons-of-the-iron-star-1200-dpi/":
        "I Sotterranei della Stella di Ferro, un complesso sotto l'abbazia dove il male si e radicato. Ogni livello scende piu in profondita nell'orrore.",

    "https://dysonlogos.blog/2020/04/15/dungeons-of-the-grand-illusionist-1200-dpi/":
        "I Sotterranei del Grande Illusionista, dove nulla e come sembra. Muri falsi, pavimenti illusori e trappole che giocano con la percezione.",

    "https://dysonlogos.blog/2020/04/08/the-ruined-keep-of-madrual/":
        "Il Fortino in Rovina di Madrual, dove le mura crollate offrono poca protezione ma molti nascondigli. Qualcuno ha reclamato queste rovine.",

    "https://dysonlogos.blog/2020/04/01/walled-temple/":
        "Il Tempio Murato, circondato da alte mura che un tempo proteggevano i fedeli dal mondo esterno. Ora proteggono il mondo da cio che e dentro.",

    "https://dysonlogos.blog/2020/03/25/the-citadel-at-sabre-lake/":
        "La Cittadella sul Lago Sabre, una fortezza costruita su un isolotto nel mezzo di un lago dalle acque taglienti come lame.",

    "https://dysonlogos.blog/2020/03/18/pharykas-walk/":
        "Il Sentiero di Pharyka, un percorso rituale che serpeggia attraverso rovine e santuari. Chi lo percorre tutto riceve una benedizione, o una maledizione.",

    "https://dysonlogos.blog/2020/03/11/the-ruins-of-boar-isle-tower-1200-dpi/":
        "Le Rovine della Torre sull'Isola del Cinghiale, dove i resti di una torre di guardia si ergono tra la vegetazione lussureggiante di un'isola selvaggia.",

    "https://dysonlogos.blog/2020/03/04/theres-a-hole-in-the-dungeon-1200-dpi/":
        "C'e un buco nel sotterraneo, e da quel buco viene qualcosa. Il pavimento crollato rivela livelli ancora piu profondi e antichi.",

    "https://dysonlogos.blog/2020/02/26/return-to-appletree-pond/":
        "Ritorno allo Stagno del Melo, un luogo idillico che nasconde un ingresso verso il sottosuolo. Le acque calme riflettono piu di quanto dovrebbero.",

    "https://dysonlogos.blog/2020/02/19/the-demon-pillars-of-iv/":
        "I Pilastri Demoniaci di Iv, colonne scolpite con volti di demoni che sembrano gridare in silenzio. L'aria tra di essi vibra di energia malvagia.",

    "https://dysonlogos.blog/2020/02/12/coopers-hole-1200-dpi/":
        "La Buca di Cooper, un sistema di caverne dove un bottaio eremita aveva il suo laboratorio. I barili abbandonati contengono piu di vino.",

    "https://dysonlogos.blog/2020/02/05/lautuntown/":
        "Lautuntown, un insediamento che e cresciuto intorno a un incrocio di strade commerciali. Un luogo di passaggio dove tutti hanno qualcosa da vendere.",

    "https://dysonlogos.blog/2020/01/29/shrine-of-the-demon-eskarna-1200-dpi/":
        "Il Santuario del Demone Eskarna, un luogo blasfemo dove i fedeli offrono sacrifici per ottenere potere proibito.",

    "https://dysonlogos.blog/2020/01/22/borderlands-caves-1200-dpi/":
        "Le Caverne delle Terre di Confine, un sistema di grotte al margine della civilta dove mostri e fuorilegge trovano rifugio.",

    "https://dysonlogos.blog/2020/01/15/somerrich-cays/":
        "Gli Isolotti di Somerrich, un arcipelago di piccole isole sabbiose dove relitti e tesori giacciono nascosti dalla marea.",

    "https://dysonlogos.blog/2020/01/08/the-false-tombs-1200-dpi/":
        "Le False Tombe, un inganno architettonico progettato per attirare i tombaroli lontano dai veri sepolcri. Ma le trappole sono reali.",

    "https://dysonlogos.blog/2020/01/01/release-the-kraken-on-axebridge-over-blackbay/":
        "Il Ponte dell'Ascia attraversa le acque nere della baia, una struttura massiccia che collega due quartieri della citta. Sotto le arcate, qualcosa si muove.",

    "https://dysonlogos.blog/2019/12/25/sanctum-of-the-blind-protean/":
        "Il Sanctum del Proteo Cieco, dove una creatura mutaforma primordiale si aggira nelle tenebre. Le pareti cambiano forma seguendo i suoi capricci.",

    "https://dysonlogos.blog/2019/12/18/bloodmarket-cave-1200-dpi/":
        "La Caverna del Mercato di Sangue, dove si commercia in merci proibite e creature vive. Il prezzo di ogni cosa si paga in rosso.",

    "https://dysonlogos.blog/2019/12/11/last-home-of-the-three-heretics-of-xaeen/":
        "L'Ultima Dimora dei Tre Eretici di Xaeen, dove tre sacerdoti banditi trovarono rifugio e continuarono i loro rituali proibiti fino alla fine.",

    "https://dysonlogos.blog/2019/12/04/bitterchains-tombs/":
        "Le Tombe di Bitterchains, catacombe dove i prigionieri incatenati venivano sepolti con le loro catene. Il tintinnio si sente ancora.",

    "https://dysonlogos.blog/2019/11/27/barrow-mounds-of-the-lich-and-famous/":
        "I Tumuli del Lich e dei Famosi, un complesso di sepolture dove potenti non-morti giacciono in un sonno irrequieto accanto ad eroi dimenticati.",

    "https://dysonlogos.blog/2019/11/20/lair-of-the-golden-wolf/":
        "La Tana del Lupo Dorato, rifugio di una bestia leggendaria il cui pelo brilla come oro. I cacciatori che la cercano raramente tornano.",

    "https://dysonlogos.blog/2019/11/13/foxtail-grotto/":
        "La Grotta di Codadivolpe, una caverna con un fiume sotterraneo dove la corrente ha scolpito forme bizzarre nella roccia.",

    "https://dysonlogos.blog/2019/11/06/dread-shrine-of-the-magi-in-sapphire/":
        "Il Santuario Terribile dei Magi in Zaffiro, dove maghi potenti compivano rituali usando gemme come catalizzatori. Le pietre blu pulsano ancora di energia.",

    "https://dysonlogos.blog/2019/10/30/pentagon-cove/":
        "La Baia del Pentagono, un'insenatura dalla forma geometrica innaturale. Le maree qui seguono ritmi che non corrispondono alla luna.",

    "https://dysonlogos.blog/2019/10/23/the-goat-shrine/":
        "Il Santuario della Capra, un luogo di culto rustico dedicato a una divinita caprina. I pastori locali lo visitano in segreto.",

    "https://dysonlogos.blog/2019/10/16/tombs-of-the-steel-makers/":
        "Le Tombe dei Forgiatori d'Acciaio, dove i maestri fabbri di un'era passata riposano circondati dalle loro creazioni migliori.",

    "https://dysonlogos.blog/2019/10/09/crowspine-labyrinth/":
        "Il Labirinto di Crowspine, un dedalo osseo costruito con le spine di creature giganti. Le pareti sono affilate e i passaggi sempre piu stretti.",

    "https://dysonlogos.blog/2019/10/02/sanvilds-delve/":
        "Lo Scavo di Sanvild, una miniera abbandonata che ha raggiunto qualcosa che non doveva essere disturbato.",

    "https://dysonlogos.blog/2019/09/25/the-ravens-rum-roosts/":
        "Il Rum e i Posatoi del Corvo, una taverna costruita in una torre dove i corvi nidificano ai piani superiori. La birra ha un retrogusto di piume.",

    "https://dysonlogos.blog/2019/09/18/isometric-tomb-of-illhan-the-binder/":
        "La Tomba di Illhan il Legatore, vista in prospettiva isometrica. I suoi sarcofagi sono incatenati e sigillati con rune di contenimento.",

    "https://dysonlogos.blog/2019/09/11/crowned-hill/":
        "La Collina Incoronata, cosi chiamata per il cerchio di pietre che ne adorna la cima. Di notte, luci misteriose danzano tra i megaliti.",

    "https://dysonlogos.blog/2019/09/04/four-pagodas-of-kwantoom/":
        "Le Quattro Pagode di Kwantoom, un complesso monastico dove quattro ordini diversi praticano discipline rivali.",

    "https://dysonlogos.blog/2019/08/28/springhollow/":
        "Vallesorgente, un insediamento bucolico in una valle protetta dove sorgenti calde sgorgano tutto l'anno.",

    "https://dysonlogos.blog/2019/08/21/the-bottomless-tombs-level-3/":
        "Il terzo livello delle Tombe Senza Fondo, dove l'architettura diventa piu elaborata e le trappole piu insidiose.",

    "https://dysonlogos.blog/2019/08/14/moonset-street-shops/":
        "I negozi di Via Moonset, una fila di botteghe dove si trova di tutto, dal comune all'impossibile. Ogni commerciante ha una storia.",

    "https://dysonlogos.blog/2019/08/07/the-warlocks-arms/":
        "Le Armi del Warlock, un'armeria specializzata in oggetti magici di dubbia provenienza. Il proprietario non fa domande.",

    "https://dysonlogos.blog/2019/07/31/gladiators-temple-dd-map/":
        "Il Tempio del Gladiatore, un'arena sacra dove i combattimenti sono rituali. I vincitori ricevono la benedizione degli dei della guerra.",

    "https://dysonlogos.blog/2019/07/24/vigilance-trail/":
        "Il Sentiero della Vigilanza, un percorso di montagna punteggiato di torri di guardia. Chi lo controlla domina il passaggio.",

    "https://dysonlogos.blog/2019/07/17/the-infested-hall/":
        "La Sala Infestata, un tempo luogo di banchetti ora invaso da creature striscianti. Le ragnatele ricoprono ogni superficie.",

    "https://dysonlogos.blog/2019/07/10/the-ruins-near-elverston-hold/":
        "Le Rovine presso la Fortezza di Elverston, resti di un insediamento distrutto dalla stessa minaccia che la fortezza doveva contenere.",

    "https://dysonlogos.blog/2019/07/03/principalities-of-black-sphinx-bay/":
        "I Principati della Baia della Sfinge Nera, una regione costiera governata da principi rivali. Intrighi e commercio marittimo dominano.",

    "https://dysonlogos.blog/2019/06/26/guddurs-wicked-teahouse/":
        "La Casa del Te Malvagia di Guddur, dove ogni infuso ha effetti imprevedibili. Il padrone di casa sorride sempre, e questo non rassicura nessuno.",

    "https://dysonlogos.blog/2019/06/19/spell-eaters-spring/":
        "La Sorgente del Divoraincantesimi, le cui acque annullano ogni magia. I maghi la temono, i guerrieri la cercano.",

    "https://dysonlogos.blog/2019/06/12/drowning-point/":
        "Il Punto dell'Annegamento, un tratto di costa dove le correnti trascinano verso il fondo. Le leggende parlano di una citta sommersa.",

    "https://dysonlogos.blog/2019/06/05/the-juicer/":
        "Lo Spremitore, un complesso meccanico di trappole progettato per ridurre gli intrusi in poltiglia. Qualcuno aveva un senso dell'umorismo macabro.",

    "https://dysonlogos.blog/2019/05/29/the-old-fort/":
        "Il Vecchio Forte, una struttura difensiva che ha visto epoche migliori. Le mura reggono ancora, ma per quanto?",

    "https://dysonlogos.blog/2019/05/22/the-old-fort-ruins-dungeon/":
        "Le rovine e i sotterranei del Vecchio Forte rivelano livelli nascosti sotto le fondamenta. La storia di questo luogo e piu profonda di quanto sembri.",

    "https://dysonlogos.blog/2019/05/15/the-twin-norkers-tagged/":
        "I Norker Gemelli, una coppia di strutture identiche che si fronteggiano. Costruite come trappola, funzionano ancora.",

    "https://dysonlogos.blog/2019/05/08/seaside-passage/":
        "Un passaggio costiero scavato dalle maree, che conduce da una spiaggia a una grotta nascosta. Praticabile solo con la bassa marea.",

    "https://dysonlogos.blog/2019/05/01/seevers-mill/":
        "Il Mulino di Seever, dove la ruota idraulica gira ancora alimentata da un torrente. Ma il mugnaio e scomparso da tempo.",

    "https://dysonlogos.blog/2019/04/24/the-holy-city-of-guerras-el-estat/":
        "La Citta Santa di Guerras-El-Estat, un labirinto di vicoli e templi dove sacro e profano si mescolano in egual misura.",

    "https://dysonlogos.blog/2019/04/17/bloodied-warrens/":
        "Le Tane Insanguinate, un sistema di tunnel dove ogni parete porta segni di artigli e macchie scure. I precedenti abitanti non se ne sono andati volontariamente.",

    "https://dysonlogos.blog/2019/04/10/beneath-the-marching-tankard/":
        "Gli avventurieri in pensione aggiungono sempre qualcosa ai locali che acquistano. Nel caso del Boccale in Marcia, Gunter Grohl ha diviso il seminterrato in due sezioni segrete.",

    "https://dysonlogos.blog/2019/04/03/the-marching-tankard-upstairs-tagged/":
        "Il piano superiore del Boccale in Marcia, con stanze per gli ospiti e un ufficio privato. L'ex avventuriero che lo gestisce tiene sempre una spada sotto il bancone.",

    "https://dysonlogos.blog/2019/03/27/the-vault-of-dahlver-nar/":
        "La Volta di Dahlver-Nar, un leggendario deposito di reliquie dove ogni oggetto e sia un dono che una maledizione.",

    "https://dysonlogos.blog/2019/03/20/barrow-mounds-of-the-lich-famous-iii/":
        "Il terzo gruppo di tumuli dove il Lich e i Famosi riposano. Ogni tumulo e una sfida diversa, e le ricompense valgono il rischio.",

    "https://dysonlogos.blog/2019/03/13/the-cinder-throne/":
        "Il Trono di Cenere, un seggio di potere forgiato nelle braci di un vulcano. Chi vi si siede comanda il fuoco, ma rischia di bruciarsi.",

    "https://dysonlogos.blog/2019/03/06/the-hydras-alehouse/":
        "La Birreria dell'Idra, dove le spine della birra sono tante quante le teste del mostro che fa da insegna. Tagliare il conto e impossibile.",

    "https://dysonlogos.blog/2019/02/27/church-of-the-oracles-in-onyx/":
        "La Chiesa degli Oracoli in Onice, dove sacerdoti veggenti leggono il futuro nelle venature della pietra nera.",

    "https://dysonlogos.blog/2019/02/20/the-lost-temple-of-aphosh-the-haunted/":
        "Il Tempio Perduto di Aphosh l'Infestato, dove il fantasma del sommo sacerdote rifiuta di abbandonare il suo santuario, secoli dopo la morte.",

    "https://dysonlogos.blog/2019/02/13/bloodied-axe-shrine/":
        "Il Santuario dell'Ascia Insanguinata, dove guerrieri offrono le loro armi in cambio della benedizione di un dio della battaglia.",

    "https://dysonlogos.blog/2019/02/06/jacobs-spur-colour-300dpi/":
        "Lo Sperone di Jacob, una formazione rocciosa che si protende nel vuoto. In cima, le rovine di un eremo offrono una vista mozzafiato e pericolosa.",

    "https://dysonlogos.blog/2019/01/30/the-dungeon-in-12-parts/":
        "Il Sotterraneo in 12 Parti, un complesso modulare dove ogni sezione puo essere esplorata indipendentemente. L'insieme e piu pericoloso della somma delle parti.",

    "https://dysonlogos.blog/2019/01/23/the-bast-inn/":
        "La Locanda di Bast, un rifugio per viaggiatori devoti alla dea felina. I gatti qui sono trattati meglio degli ospiti.",

    "https://dysonlogos.blog/2019/01/16/royal-catacombs-of-adrih/":
        "Le Catacombe Reali di Adrih, dove re e regine riposano in sarcofagi dorati. Le guardie reali, non-morte, vegliano ancora.",

    "https://dysonlogos.blog/2019/01/09/pillar-of-the-igesej-loremaster/":
        "Il Pilastro del Maestro del Sapere Igesej, una torre isometrica dove la conoscenza e conservata in forme architettoniche.",

    "https://dysonlogos.blog/2019/01/02/kabrels-tower/":
        "La Torre di Kabrel, dimora di un mago solitario che preferisce la compagnia dei libri a quella degli umani.",

    "https://dysonlogos.blog/2018/12/26/uogralas-city-of-the-frogs/":
        "Uogralas, la Citta delle Rane, un insediamento anfibio dove le costruzioni emergono dall'acqua e le strade sono canali.",

    "https://dysonlogos.blog/2018/12/19/hurren-city-of-the-elders/":
        "Hurren, la Citta degli Anziani, governata da un consiglio di vegliardi la cui saggezza e pari solo alla loro diffidenza verso gli stranieri.",

    "https://dysonlogos.blog/2018/12/12/the-blind-lamias-cave/":
        "La Caverna della Lamia Cieca, dove una creatura privata della vista ha affinato altri sensi mortali. Ogni passo rischia di tradire la vostra presenza.",

    "https://dysonlogos.blog/2018/12/05/iyesgarten-regional-map/":
        "Mappa regionale di Iyesgarten, una regione rurale punteggiata di villaggi e fattorie dove la vita semplice nasconde pericoli insospettati.",

    "https://dysonlogos.blog/2018/11/28/the-iyesgarten-inn/":
        "La Locanda di Iyesgarten, cuore pulsante del villaggio dove i viaggiatori trovano ristoro e gli avventurieri trovano lavoro.",

    "https://dysonlogos.blog/2018/11/21/the-village-of-iyesgarten/":
        "Il Villaggio di Iyesgarten, un pugno di case attorno a una piazza dove il tempo sembra scorrere piu lento.",

    "https://dysonlogos.blog/2018/11/14/ssa-tuns-lake-of-milk/":
        "Il Lago di Latte di Ssa-Tun, le cui acque bianche e opache nascondono il fondo e qualunque cosa vi dimori.",

    "https://dysonlogos.blog/2018/11/07/spectres-tower/":
        "La Torre dello Spettro, abbandonata dopo che il suo ultimo abitante non ha mai veramente lasciato. Le luci che si vedono di notte non sono candele.",

    "https://dysonlogos.blog/2018/10/31/the-temple-walk/":
        "La Passeggiata del Tempio, un percorso processionale fiancheggiato da statue di divinita che conduce al sanctum. Ogni statua osserva chi passa.",

    "https://dysonlogos.blog/2018/10/24/chambers-of-the-absent-city/":
        "Le Camere della Citta Assente, resti sotterranei di una citta che e scomparsa dalla superficie. Solo queste sale testimoniano la sua esistenza.",

    "https://dysonlogos.blog/2018/10/17/shrine-of-the-emperor-of-bones/":
        "Il Santuario dell'Imperatore delle Ossa, dove un sovrano non-morto riceve l'adorazione di fedeli che preferiscono la morte alla vita.",

    "https://dysonlogos.blog/2018/10/10/the-behkon-inn/":
        "La Locanda di Behkon, un edificio modesto ma accogliente sulla strada principale. L'oste ha una memoria prodigiosa per i volti.",

    "https://dysonlogos.blog/2018/10/03/sewer-connectors/":
        "Raccordi fognari che collegano diverse sezioni del sistema sotterraneo. Moduli componibili per espandere qualsiasi rete fognaria.",

    "https://dysonlogos.blog/2018/09/26/lady-whites-ruins/":
        "Le Rovine di Lady White, resti della dimora di una nobildonna la cui storia e avvolta nel mistero e nel dolore.",

    "https://dysonlogos.blog/2018/09/19/isle-of-kheyus-colour/":
        "L'Isola di Kheyus, un gioiello tropicale con spiagge dorate e una giungla impenetrabile che nasconde templi perduti.",

    "https://dysonlogos.blog/2018/09/12/summerthorpe/":
        "Summerthorpe, un villaggio estivo dove le giornate sono lunghe e i segreti sono sepolti poco sotto la superficie.",

    "https://dysonlogos.blog/2018/09/05/labhruinns-tavern/":
        "La Taverna di Labhruinn, gestita da un ex avventuriero le cui cicatrici raccontano piu storie delle canzoni dei bardi.",

    "https://dysonlogos.blog/2018/08/29/an-nayyirs-pyramid/":
        "La Piramide di An-Nayyir, un monumento funebre nel deserto dove i sacerdoti del sole seppellivano i loro piu grandi campioni.",

    "https://dysonlogos.blog/2018/08/22/defiled-waters/":
        "Le Acque Profanate, un sistema di caverne acquatiche dove un rituale blasfemo ha corrotto le sorgenti. Chi beve quest'acqua non e piu lo stesso.",

    "https://dysonlogos.blog/2018/08/15/cage-street-sewers/":
        "Le Fogne di Via Cage, tunnel sotto il quartiere piu malfamato della citta. I criminali li conoscono meglio delle guardie.",

    "https://dysonlogos.blog/2018/08/08/the-dark-caverns-of-turr/":
        "Le Caverne Oscure di Turr, un sistema di grotte dove l'oscurita sembra avere una consistenza quasi fisica. Le torce bruciano a meta potenza.",

    "https://dysonlogos.blog/2018/08/01/sewer-elements/":
        "Elementi fognari modulari per costruire reti di tunnel sotterranei. Ogni pezzo si incastra per creare complessi unici.",

    "https://dysonlogos.blog/2018/07/25/temple-crypts-of-the-wraith-priests/":
        "Le Cripte del Tempio dei Sacerdoti Spettrali, dove i ministri del culto continuano a officiare i loro riti anche dopo la morte.",

    "https://dysonlogos.blog/2018/07/18/the-stone-trolls-lantern/":
        "La Lanterna del Troll di Pietra, una caverna dove un troll pietrificato regge una lanterna magica che non si spegne mai.",

    "https://dysonlogos.blog/2018/07/11/west-sewers-complex/":
        "Il Complesso Fognario Ovest, una rete di tunnel e cisterne sotto il quartiere occidentale. Le acque reflue nascondono passaggi segreti.",

    "https://dysonlogos.blog/2018/07/04/psychedelic-cellar-of-the-stone-giants/":
        "La Cantina Psichedelica dei Giganti di Pietra, dove funghi allucinogeni giganti crescono tra le botti. I giganti li usano per le loro visioni.",

    "https://dysonlogos.blog/2018/06/27/south-sewers-hideout/":
        "Il Nascondiglio delle Fogne Sud, un rifugio clandestino ricavato in una cisterna abbandonata. Perfetto per chi non vuole essere trovato.",

    "https://dysonlogos.blog/2018/06/20/smugglers-lodge/":
        "Il Rifugio dei Contrabbandieri, nascosto in una rete di caverne costiere. I tunnel collegano il mare alla citta senza passare per le porte.",

    "https://dysonlogos.blog/2018/06/13/gauntlet-of-the-flintcrowned-ghouls/":
        "Il Guanto di Sfida dei Ghoul dalla Corona di Selce, un percorso mortale dove i non-morti incoronati mettono alla prova chiunque osi passare.",

    "https://dysonlogos.blog/2018/06/06/shieldricks-tower-inn/":
        "La Locanda della Torre di Shieldrick, costruita dentro e intorno a un'antica torre di guardia. Le stanze ai piani alti offrono viste panoramiche.",

    "https://dysonlogos.blog/2018/05/30/rosewood-street-sewers/":
        "Le Fogne di Via Rosewood, un tratto di tunnel sorprendentemente ben costruito sotto il quartiere nobile. Anche i ricchi hanno segreti da nascondere.",

    "https://dysonlogos.blog/2018/05/23/the-red-descent/":
        "La Discesa Rossa, un santuario verticale scavato nella roccia che scende attraverso caverne e templi verso profondita sconosciute.",

    "https://dysonlogos.blog/2018/05/16/the-old-turnip-inn/":
        "La Locanda della Vecchia Rapa, un posto senza pretese ma dal cibo onesto. Il nome deriva dall'enorme rapa scolpita che funge da insegna.",

    "https://dysonlogos.blog/2018/05/09/the-phoenix-diadem/":
        "Il Diadema della Fenice, un artefatto custodito in una camera segreta. La corona ardente promette resurrezione ma esige un prezzo terribile.",

    "https://dysonlogos.blog/2018/05/02/guild-temple/":
        "Il Tempio della Gilda, dove commercianti e artigiani si riuniscono sotto la protezione di una divinita del commercio. Gli affari qui sono sacri.",

    "https://dysonlogos.blog/2018/04/25/crypt-of-the-smith/":
        "La Cripta del Fabbro, dove un leggendario artigiano fu sepolto con i suoi strumenti e le sue creazioni migliori. L'incudine risuona ancora.",

    "https://dysonlogos.blog/2018/04/18/the-ruins-of-greymail-clanhold/":
        "Le Rovine della Fortezza del Clan Greymail, un tempo sede di un clan nanico orgoglioso. Il drago che li ha sconfitti potrebbe ancora essere nei paraggi.",

    "https://dysonlogos.blog/2018/04/11/the-vanshiro-reliquary/":
        "Il Reliquiario di Vanshiro, una camera blindata dove reliquie di potere sono conservate e protette da guardiani instancabili.",

    "https://dysonlogos.blog/2018/04/04/onyx-hill-ruins/":
        "Le Rovine della Collina di Onice, dove pietre nere emergono dal terreno come denti rotti. Un luogo evitato da tutti tranne che dai piu disperati.",

    "https://dysonlogos.blog/2018/03/28/the-half-cask-tavern/":
        "La Taverna del Mezzo Barile, cosi piccola che un barile tagliato a meta serve da bancone. Ma la birra e sorprendentemente buona.",

    "https://dysonlogos.blog/2018/03/21/the-court-of-summer-wines/":
        "La Corte dei Vini Estivi, un palazzo dove il vino scorre da fontane e i piaceri sono l'unica legge. Un luogo di decadenza e pericolo.",

    "https://dysonlogos.blog/2018/03/14/letath/":
        "Letath, un insediamento misterioso dove gli abitanti parlano poco e gli stranieri sono osservati con sospetto.",

    "https://dysonlogos.blog/2018/03/07/oubliette-of-the-forgotten-magus/":
        "L'Oubliette del Mago Dimenticato, una prigione magica dove un incantatore fu rinchiuso e poi rimosso dalla memoria del mondo.",

    "https://dysonlogos.blog/2018/02/28/ashryn-spire/":
        "La Guglia di Ashryn, una torre cosi alta e sottile da sembrare impossibile. Ai piani superiori, il vento canta melodie arcane.",

    "https://dysonlogos.blog/2018/02/21/the-fire-beetle-ale-house/":
        "La Birreria dello Scarabeo di Fuoco, illuminata dalla bioluminescenza degli insetti che le danno il nome. La birra brilla al buio.",

    "https://dysonlogos.blog/2018/02/14/robrus-beach-cave/":
        "La Caverna della Spiaggia di Robrus, accessibile solo con la bassa marea. Le pareti luccicano di cristalli di sale.",

    "https://dysonlogos.blog/2018/02/07/control/":
        "Controllo, una struttura enigmatica il cui scopo originale si e perso nel tempo. I pannelli alle pareti sembrano aspettare comandi.",

    "https://dysonlogos.blog/2018/01/31/the-blessed-monastery/":
        "Il Monastero Benedetto, un rifugio di pace e preghiera circondato da giardini meticolosi. Ma la benedizione ha un prezzo.",

    "https://dysonlogos.blog/2018/01/24/the-bubble-city-of-oublos/":
        "La Citta delle Bolle di Oublos, un insediamento fantastico dove le strutture sono sfere traslucide. La fisica qui segue regole diverse.",

    "https://dysonlogos.blog/2018/01/17/the-bottomless-tombs-level-2/":
        "Il secondo livello delle Tombe Senza Fondo, dove l'architettura si fa piu complessa e le trappole piu sofisticate.",

    "https://dysonlogos.blog/2018/01/10/twelve-goats-tavern/":
        "La Taverna dei Dodici Capri, dove dodici teste di capra imbalsamate decorano le pareti. Ognuna, dicono, racconta una storia diversa.",

    "https://dysonlogos.blog/2018/01/03/black-armoury-of-the-mad-king/":
        "L'Armeria Nera del Re Folle, dove armi maledette e armature possedute attendono nuovi portatori. Il re era pazzo, ma le sue armi sono reali.",

    "https://dysonlogos.blog/2017/12/27/will-o-the-wisp/":
        "Fuoco Fatuo, un luogo dove luci ingannevoli guidano i viandanti nella palude. Chi le segue non torna mai indietro.",

    "https://dysonlogos.blog/2017/12/20/sanhelter-keep/":
        "Il Fortino di Sanhelter, un presidio militare in una posizione strategica. Le sue mura sottili compensano con l'astuzia del progetto.",

    "https://dysonlogos.blog/2017/12/13/the-wooden-duck-inn/":
        "La Locanda dell'Anatra di Legno, famosa per la sua insegna scolpita e per lo stufato che nessuno sa di cosa sia fatto.",

    "https://dysonlogos.blog/2017/12/06/the-vault-of-tranquility/":
        "La Volta della Tranquillita, una camera protetta dove il silenzio e assoluto. Un rifugio perfetto, se non fosse per cio che il silenzio nasconde.",

    "https://dysonlogos.blog/2017/11/29/the-ruins-of-charnesse/":
        "Le Rovine di Charnesse, vestigia di una citta distrutta da una calamita dimenticata. Le fondamenta raccontano di grandezza perduta.",

    "https://dysonlogos.blog/2017/11/22/yruvex-swamps/":
        "Le Paludi di Yruvex, un pantano impenetrabile dove la nebbia non si alza mai e le creature che vi abitano non sono di questo mondo.",

    "https://dysonlogos.blog/2017/11/15/the-banshees-tower/":
        "La Torre della Banshee, dove il grido di uno spirito tormentato echeggia notte dopo notte. I vicini hanno imparato a dormire con le orecchie tappate.",

    "https://dysonlogos.blog/2017/11/08/the-tower-faced-demon/":
        "Il Demone dalla Faccia di Torre, una struttura che sembra un edificio ma e il corpo pietrificato di un demone colossale.",

    "https://dysonlogos.blog/2017/11/01/lesser-temple-of-the-heretics/":
        "Il Tempio Minore degli Eretici, un santuario clandestino dove dottrine proibite vengono insegnate a pochi iniziati.",

    "https://dysonlogos.blog/2017/10/25/temple-of-the-seven-heretics/":
        "Il Tempio dei Sette Eretici, dove sette sacerdoti rinnegati fondarono un culto che sfidava gli dei stessi. Le loro statue sorridono ancora.",

    "https://dysonlogos.blog/2017/10/18/brenovale-castle/":
        "Il Castello di Brenovale, una fortezza che ha cambiato proprietario decine di volte. Ogni nuovo signore ha aggiunto stanze e segreti.",

    "https://dysonlogos.blog/2017/10/11/loreans-manor/":
        "Il Maniero di Lorean, una dimora elegante dove la raffinatezza nasconde scopi oscuri. Le cene qui sono memorabili, per chi sopravvive.",

    "https://dysonlogos.blog/2017/10/04/paradise-control/":
        "Controllo del Paradiso, una struttura misteriosa che sembra regolare qualcosa di cosmico. I pannelli di controllo brillano ancora.",

    "https://dysonlogos.blog/2017/09/27/the-portal-nexus-with-grid/":
        "Il Nodo dei Portali, un crocevia dimensionale dove porte magiche conducono a luoghi sconosciuti. Ogni portale ha un prezzo.",

    "https://dysonlogos.blog/2017/09/20/rose-point-manor/":
        "Il Maniero di Rose Point, una residenza costiera dove il profumo delle rose si mescola alla brezza marina. La bellezza nasconde spine.",

    "https://dysonlogos.blog/2017/09/13/beneath-rose-point-manor/":
        "Sotto il Maniero di Rose Point si estendono cantine e passaggi segreti che conducono alle grotte sulla scogliera.",

    "https://dysonlogos.blog/2017/09/06/vault-of-the-blue-golem/":
        "La Volta del Golem Blu, dove un costrutto di cristallo azzurro custodisce tesori da secoli. Non ha mai fallito il suo compito.",

    "https://dysonlogos.blog/2017/08/30/strange-ruins/":
        "Rovine Misteriose, resti di una struttura la cui architettura non corrisponde a nessuna civilta conosciuta.",

    "https://dysonlogos.blog/2017/08/23/the-many-chambers-of-izzets-folly/":
        "Le Molte Camere della Follia di Izzet, un complesso labirintico costruito da un mago le cui ambizioni superavano la sua sanita mentale.",

    "https://dysonlogos.blog/2017/08/16/cliffstable-on-kerstal/":
        "Cliffstable su Kerstal, un insediamento aggrappato al bordo di una scogliera. Le case pendono sul vuoto e le scale sono l'unica via.",

    "https://dysonlogos.blog/2017/08/09/hollowstone-bandit-camp/":
        "L'Accampamento dei Banditi di Hollowstone, costruito su una prominenza rocciosa nei boschi. L'elfo bandito Illsong e i suoi compagni lo stanno trasformando in una vera fortezza.",

    "https://dysonlogos.blog/2017/08/02/the-arcane-waters/":
        "Le Acque Arcane, un sistema di grotte dove l'acqua e intrisa di magia. Chi vi nuota sente voci e vede visioni.",

    "https://dysonlogos.blog/2017/07/26/brentil-tower/":
        "La Torre di Brentil, una struttura difensiva che controlla un passo montano. Il suo guardiano non riceve visitatori da anni.",

    "https://dysonlogos.blog/2017/07/19/palace-of-the-sands/":
        "Il Palazzo delle Sabbie, una residenza nel deserto che il vento scopre e ricopre secondo i suoi capricci. Trovarla e gia un'avventura.",

    "https://dysonlogos.blog/2017/07/12/dungeon-of-the-third-eye/":
        "Il Sotterraneo del Terzo Occhio, un complesso dedicato alla divinazione e alla visione interiore. Le pareti sono occhi che osservano.",

    "https://dysonlogos.blog/2017/07/05/princes-harbour-map-3/":
        "La terza sezione del Porto del Principe, dove i magazzini portuali nascondono piu di semplici merci.",

    "https://dysonlogos.blog/2017/06/28/stony-hill/":
        "La Collina Rocciosa, un'altura dove rocce affioranti e rovine antiche creano un terreno perfetto per imboscate.",

    "https://dysonlogos.blog/2017/06/21/old-cruik-hollow/":
        "La Vecchia Conca di Cruik, una depressione nel terreno dove la nebbia si raccoglie e le storie di fantasmi abbondano.",

    "https://dysonlogos.blog/2017/06/14/white-crag-fortress/":
        "La Fortezza della Rupe Bianca, scolpita nel calcare di una scogliera imponente. Le sue mura bianche brillano al sole come un faro.",

    "https://dysonlogos.blog/2017/06/07/princes-harbour-map-2/":
        "La seconda sezione del Porto del Principe, il cuore commerciale dove navi da ogni dove scaricano merci esotiche.",

    "https://dysonlogos.blog/2017/05/31/greths-island-keep/":
        "Il Fortino sull'Isola di Greth, una piccola fortezza su un'isola rocciosa. Raggiungibile solo via mare, e un rifugio ideale o una prigione perfetta.",

    "https://dysonlogos.blog/2017/05/24/tomb-of-the-kirin-born-prince/":
        "La Tomba del Principe Nato dal Kirin, un sepolcro regale dove un principe benedetto da una creatura celestiale riposa circondato da tesori divini.",

    "https://dysonlogos.blog/2017/05/17/wyvernseeker-rock/":
        "La Roccia di Wyvernseeker, un picco isolato dove i cacciatori di viverne piazzano le loro trappole. La vista dalla cima e magnifica, se si sopravvive.",

    "https://dysonlogos.blog/2017/05/10/priors-hill/":
        "La Collina del Priore, dove un monastero in rovina domina la vallata. Il priore e scomparso ma la sua campana suona ancora da sola.",

    "https://dysonlogos.blog/2017/05/03/princes-harbour-map-1/":
        "La prima sezione del Porto del Principe, dove moli e banchine accolgono navi da tutto il mondo conosciuto.",

    "https://dysonlogos.blog/2017/04/26/redrock-cays/":
        "Gli Isolotti di Redrock, formazioni di arenaria rossa che emergono dal mare. Le grotte nelle scogliere custodiscono tesori di pirati.",

    "https://dysonlogos.blog/2017/04/19/circle-crypts-of-the-ophidian-emperor/":
        "Le Cripte Circolari dell'Imperatore Serpente, un complesso funebre a pianta circolare dove un sovrano rettiliano giace con i suoi servitori.",

    "https://dysonlogos.blog/2017/04/12/hevlod-manor/":
        "Il Maniero di Hevlod, una dimora di campagna il cui proprietario ha gusti eccentrici. Le stanze sono piene di curiosita e pericoli.",

    "https://dysonlogos.blog/2017/04/05/briar-keep-2018-edition/":
        "Il Fortino di Briar, avvolto da rovi impenetrabili che sembrano avere una volonta propria. Chi entra deve farsi strada col fuoco o con la magia.",

    "https://dysonlogos.blog/2017/03/29/drow-spire-fortress/":
        "La Fortezza della Guglia Drow, una struttura vertiginosa costruita dai drow nelle profondita della terra. Le sue difese sono tanto eleganti quanto letali.",
}


def main():
    # Load raw scrape (same order as DESCRIPTIONS dict)
    with open("dyson_raw.json", encoding="utf-8") as f:
        raw = json.load(f)

    # Re-key descriptions by actual source_url from raw data (positional match)
    desc_values = list(DESCRIPTIONS.values())
    desc_by_url = {}
    for i, r in enumerate(raw):
        if i < len(desc_values):
            desc_by_url[r["source_url"]] = desc_values[i]

    # Load current mappe.json
    mappe_path = DATA_DIR / "mappe.json"
    with open(mappe_path, encoding="utf-8") as f:
        mappe = json.load(f)

    # Match descriptions to mappe entries by url_originale
    matched = 0
    missing = 0
    for m in mappe:
        url = m.get("url_originale", "")
        if url in desc_by_url:
            m["descrizione"] = desc_by_url[url]
            matched += 1
        else:
            m["descrizione"] = ""
            if url:
                missing += 1

    with open(mappe_path, "w", encoding="utf-8") as f:
        json.dump(mappe, f, indent=2, ensure_ascii=False)
        f.write("\n")

    print(f"Matched {matched}/{len(mappe)} descriptions ({missing} missing)")


if __name__ == "__main__":
    main()
