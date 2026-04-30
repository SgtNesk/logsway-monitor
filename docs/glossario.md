# Glossario Tecnico

Termini usati in Logsway e nella documentazione.

---

**Agent**
Il programma che installi su ogni macchina che vuoi monitorare. Legge CPU, RAM, disco e invia i dati al server.

**API**
Application Programming Interface. Il modo in cui i programmi si parlano. L'agent usa l'API del server per inviare le metriche. URL base: `/api/v1/`

**Binary / Binario**
Un file eseguibile già compilato. Non serve installare linguaggi di programmazione: scarichi il file, lo esegui.

**CPU**
Central Processing Unit. Il "cervello" del computer. Logsway mostra quanta percentuale della CPU è in uso. Al 95% → stato critico.

**curl**
Strumento a riga di comando per fare richieste HTTP. Usato per scaricare file e testare connessioni. Esempio: `curl http://server:8080/api/v1/health`

**Dashboard**
La pagina principale di Logsway. Mostra tutti gli host con il loro stato in tempo reale.

**Disco / Disk**
La percentuale di spazio su disco usato. Include tutte le partizioni montate.

**Firewall**
Sistema che blocca o permette le connessioni di rete. Se l'agent non raggiunge il server, controlla il firewall. Su Ubuntu si gestisce con `ufw`.

**Host**
Una macchina (server, VM, container) monitorata da Logsway. Appare nella dashboard con nome, stato e metriche.

**HTTP**
HyperText Transfer Protocol. Il protocollo usato da browser e API web. Logsway usa HTTP sulla porta 8080.

**IP / Indirizzo IP**
L'indirizzo numerico di una macchina in rete. Esempio: `192.168.1.10`. Serve per configurare la connessione agent→server.

**Load Average**
Misura di quante operazioni il sistema ha in coda. Valori tipici per un sistema sano: sotto il numero di core CPU. Es: su una macchina a 4 core, load average 3.5 è normale.

**Log**
File di testo dove un programma scrive quello che sta facendo e gli eventuali errori. I log di Logsway si trovano in `/var/log/logsway/`.

**Metriche**
Valori numerici che descrivono lo stato di un sistema. CPU, RAM, disco, load average sono tutti esempi di metriche.

**Porta / Port**
Un numero che identifica un servizio specifico su una macchina. Logsway usa la porta `8080`. Il browser usa la porta `80` per HTTP e `443` per HTTPS.

**RAM**
Random Access Memory. La memoria operativa del computer. Diverso dal disco: la RAM è temporanea, veloce, costosa. Logsway mostra quanta è in uso.

**Root**
L'utente amministratore su Linux. Ha accesso completo al sistema. Gli installer di Logsway richiedono root (`sudo`).

**/nongreen**
La pagina "Problems" di Logsway (ispirata a Xymon). Mostra solo gli host con problemi: warning, critical o offline.

**SSH**
Secure Shell. Protocollo per collegarsi in modo sicuro a un computer remoto via terminale. Esempio: `ssh admin@192.168.1.10`

**Systemd**
Il sistema di gestione dei servizi su Linux moderno. Logsway si installa come servizio systemd. Comandi principali: `systemctl start/stop/restart/status logsway`

**Tag**
Etichetta testuale che puoi assegnare a un host nell'agent config. Serve per raggruppare e filtrare gli host nella dashboard.

**Threshold / Soglia**
Valore limite oltre il quale una metrica cambia stato. Esempio: CPU warning threshold = 80% significa che sopra l'80% l'host diventa giallo.

**YAML**
Yet Another Markup Language. Il formato usato dai file di configurazione di Logsway. Usa l'indentazione per la struttura. Attenzione: usa SPAZI, non TAB.

**Xymon**
Tool di monitoring open source degli anni 2000. Logsway ne è ispirato graficamente: status colorati per host, pagina `/nongreen` per i problemi.
