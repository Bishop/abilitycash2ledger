# AbilityCash to Ledger converter

Some tools to convert [AbilityCash](https://dervish.ru/) database
to [Plain text format](https://plaintextaccounting.org/).

Historically supports XML, CSV and SQLite directly.

Prepare Excel export to import:

```sh
in2csv --sheet "Transactions" abilitycash/source.xlsx > abilitycash/txs.csv
in2csv --sheet "Account plans" abilitycash/source.xlsx > abilitycash/structure.csv
in2csv --sheet "Accounts" abilitycash/source.xlsx > abilitycash/accounts.csv
in2csv --sheet "Rates" abilitycash/source.xlsx > abilitycash/rates.csv
```
