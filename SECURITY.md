# Security Policy

Source Network appreciates investigative work into security vulnerabilities carried out by well-intentioned, ethical security researchers. We follow the practice of [responsible disclosure](https://en.wikipedia.org/wiki/Coordinated_vulnerability_disclosure) in order to best protect Source's user-base from the impact of security issues. This means: We will respond to security incidents as a priority. We will fix the issue as soon as is practical, keeping in mind that not all risks are created equal. We will always transparently let the community know about any incident that affects them.

On receipt the security team will: Review the report, verify the vulnerability and respond with confirmation and/or further information requests; we typically reply within 24 hours. Once the reported security bug has been addressed we will notify the Researcher, who is then welcome to optionally disclose publicly.

We maintain a Hall of Fame to recognize those who have responsibly disclosed security issues to us in the past.


## Supported Versions

TBD which versions of your project are currently being supported with security updates.

| Version | Supported          |
| ------- | ------------------ |
| 5.1.x   | :white_check_mark: |
| 5.0.x   | :x:                |
| 4.0.x   | :white_check_mark: |
| < 4.0   | :x:                |


## Audit reports

TBD


## Reporting a Vulnerability

TBD how to report a vulnerability: ..., where to go, how often they can expect to get an update on a
reported vulnerability, what to expect if the vulnerability is accepted or declined, etc

TBD key


## Disclosure Process

TBD


## Example Vulnerabilities

### Specification

* Conceptual flaws
* Ambiguities, inconsistencies, or incorrect statements
* Mis-match between specification and implementation of any component

### Networking and RPC

* Write-access to anything besides sending transactions
* Denial-of-Service
    * Amplification attacks
    * Resource abuse
    * Deadlocks and race conditions
* Leakage of secrets

### Libraries

* Serialization
* Reading/Writing files and databases

### Cryptography

* Signatures
* Hash algorithms and Merkle trees for block validation
* Authenticated encryption for P2P connections