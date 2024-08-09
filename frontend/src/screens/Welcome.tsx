import React from "react";
import { Link, useNavigate } from "react-router-dom";
import Container from "src/components/Container";
import { Button } from "src/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "src/components/ui/dialog";
import { useInfo } from "src/hooks/useInfo";

export function Welcome() {
  const { data: info } = useInfo();
  const navigate = useNavigate();

  React.useEffect(() => {
    if (!info?.setupCompleted) {
      return;
    }
    navigate("/");
  }, [info, navigate]);

  return (
    <Container>
      <div className="grid text-center gap-5">
        <div className="grid gap-2">
          <h1 className="font-semibold text-2xl font-headline">
            Welcome to Alby Hub
          </h1>
          <p className="text-muted-foreground">
            A powerful, all-in-one bitcoin lightning wallet with the superpower
            of connecting to applications.
          </p>
        </div>
        <div className="grid gap-2">
          <Link
            to={
              info?.backendType
                ? "/setup/password?node=preset" // node already setup through env variables
                : "/setup/password?node=ldk"
            }
            className="w-full"
          >
            <Button className="w-full">
              Get Started
              {info?.backendType && ` (${info?.backendType})`}
            </Button>
          </Link>

          {info?.enableAdvancedSetup && (
            <Link to="/setup/advanced" className="w-full">
              <Button variant="secondary" className="w-full">
                Advanced Setup
              </Button>
            </Link>
          )}
        </div>
        <div className="text-sm text-muted-foreground">
          By continuing, you agree to our <br />
          <Dialog>
            <DialogTrigger asChild>
              <a className="underline cursor-pointer">
                Software License Agreement
              </a>
            </DialogTrigger>
            <DialogContent className="lg:max-w-screen-lg overflow-y-scroll max-h-[90%]">
              <DialogHeader>
                <DialogTitle>End User License Agreement</DialogTitle>
              </DialogHeader>
              <div className="flex flex-col gap-2 text-sm">
                <h2 className="font-semibold">Introduction</h2>
                <p>
                  Please read this end user license agreement carefully before
                  using this application. By using this application, you signify
                  your assent to and acceptance of this end user license
                  agreement and acknowledge you have read and understand the
                  terms. An individual acting on behalf of an entity represents
                  that he or she has the authority to enter into this end user
                  license agreement on behalf of the entity. This end user
                  license agreement does not provide any rights to Alby services
                  such as software maintenance, upgrades or support.
                </p>
                <p>
                  This end user license agreement (“EULA”) governs the use of
                  the application that includes or refers to this license and
                  any related updates, source code, appearance, structure and
                  organization, regardless of the delivery mechanism.
                </p>
                <h2 className="font-semibold">
                  Interpretation and Definitions
                </h2>
                <p>
                  The words of which the initial letter is capitalized have
                  meanings defined under the following conditions. The following
                  definitions shall have the same meaning regardless of whether
                  they appear in singular or in plural.
                </p>
                <p>
                  For the purposes of this End-User License Agreement:
                  <ul className="list-disc list-inside">
                    <li>
                      <b>Agreement</b> means this End-User License Agreement
                      that forms the entire agreement between You and the
                      Company regarding the use of the Application.
                    </li>
                    <li>
                      <b>Application</b> means the software program provided by
                      the Company downloaded by You to a Device.{" "}
                    </li>
                    <li>
                      <b>Application Store</b> means the digital distribution
                      service operated and developed by Apple Inc. (Apple App
                      Store), Google Inc. (Google Play Store) or others by which
                      the Application has been downloaded to your Device.
                    </li>
                    <li>
                      <b>Company</b> (referred to as either "the Company", "We",
                      “Alby”, “Alby Inc.” "Us" or "Our" in this Agreement)
                      refers to Alby.
                    </li>
                    <li>
                      <b>Content</b> refers to content such as text, images, or
                      other information that can be posted, uploaded, linked to
                      or otherwise made available by You, regardless of the form
                      of that content.
                    </li>
                    <li>
                      <b>Device</b> means any device that can run the
                      Application such as a server, a computer or a digital
                      tablet.
                    </li>
                    <li>
                      <b>Third-Party Services</b> means any services or content
                      (including data, information, applications and other
                      products services) provided by a third-party that may be
                      displayed, included or made available by the Application.
                    </li>
                    <li>
                      <b>You</b> means the individual accessing or using the
                      Application or the company, or other legal entity on
                      behalf of which such individual is accessing or using the
                      Application, as applicable.
                    </li>
                  </ul>
                </p>
                <h2 className="font-semibold">Acknowledgment</h2>
                <p>
                  By using the Application, You affirm that you are the older of
                  18 years old or the age of majority as required by your local
                  law and have the capacity to enter into this Agreement. If you
                  are accessing the Application on behalf of the company you
                  work for, you also affirm you have the proper grant of
                  authority and capacity to enter into this Agreement on behalf
                  of such company. You are agreeing to be bound by the terms and
                  conditions of this Agreement. If You do not agree to the terms
                  of this Agreement, do not use the Application. This Agreement
                  is a legal document between You and Alby and it governs your
                  use of the Application made available to You by Alby. The
                  Application is licensed, not sold, to You by Alby for use
                  strictly in accordance with the terms of this Agreement.
                </p>
                <h2 className="font-semibold">License</h2>
                <p>
                  Subject to the following terms, Alby grants to you a
                  perpetual, worldwide license to the Application pursuant to
                  the Apache-2.0 license.
                  (https://github.com/getAlby/hub?tab=Apache-2.0-1-ov-file#readme).
                  This EULA pertains solely to the Application and does not
                  limit your rights under, or grant you rights that supersede,
                  the license terms of any particular component.
                </p>
                <h2 className="font-semibold">Rights and Responsibilities</h2>
                <p>
                  This Application functions as a free, open source application.
                  The Application does not constitute an account where We or
                  other third parties serve as financial intermediaries or
                  custodians of Your assets.
                </p>
                <p>
                  With respect to the Application, Alby does not receive or
                  store the unlock password, nor any recovery phrase or private
                  keys, or the individual transaction history. Alby cannot
                  assist you with password retrieval. You are solely responsible
                  for remembering, storing and keeping secret the password and
                  recovery phrase. The assets you have associated with the
                  Application may become inaccessible if you do not know or keep
                  secret your password, private keys and recovery phrase. Any
                  third party with knowledge of one or more of your credentials
                  (including, without limitation, a recovery phrase or password)
                  can access the assets via your Application or recover your
                  funds in another app and initiate transactions.
                </p>
                <p>
                  You agree to take responsibility for all activities that occur
                  with the application operated by you and accept all risks of
                  any authorized or unauthorized access to your application, to
                  the maximum extent permitted by law.
                </p>
                <p>
                  All transaction requests are irreversible. Alby shall not be
                  liable to you for any malformed transaction payloads created
                  using the Application. Once transaction details have been
                  submitted to the asset network, we cannot assist you in
                  canceling or otherwise modifying the transaction or
                  transaction details. Alby has no control over your assets
                  stored by the Application and does not have the ability to
                  facilitate any cancellation or modification requests.
                  Furthermore Alby has no control over the data that you provide
                  to form any transaction using the Application; or how
                  transactions are processed in the asset network. Alby
                  therefore cannot and does not ensure that any transaction
                  details you submit via the Application will be confirmed on
                  the asset network. The transaction details submitted by you
                  via the Application provided by us may not be completed, or
                  may be substantially delayed, by the asset network used to
                  process the transaction. We do not guarantee that the
                  Application can transfer title or right in the asset or make
                  any warranties whatsoever with regard to title.
                </p>
                <h2 className="font-semibold">Intellectual Property</h2>
                <p>
                  We retain all right, title, and interest in all of Alby's
                  brands, logos, and trademarks, including, but not limited to
                  and variations of the wording of the aforementioned brands,
                  logos, and trademarks. This EULA does not permit you to
                  distribute the Programs using the Company's trademarks,
                  regardless of whether the Application has been modified.
                </p>

                <h2 className="font-semibold">
                  Modifications to the Application
                </h2>
                <p>
                  The Company reserves the right to modify, suspend or
                  discontinue, temporarily or permanently, the Application or
                  any service to which it connects, with or without notice and
                  without liability to You.
                </p>
                <h2 className="font-semibold">Updates to the Application</h2>
                <p>
                  The Company may from time to time provide enhancements or
                  improvements to the features/functionality of the Application,
                  which may include patches, bug fixes, updates, upgrades and
                  other modifications. Updates may modify or delete certain
                  features and/or functionalities of the Application. You agree
                  that the Company has no obligation to (i) provide any Updates,
                  or (ii) continue to provide or enable any particular features
                  and/or functionalities of the Application to You. You further
                  agree that all updates or any other modifications will be (i)
                  deemed to constitute an integral part of the Application, and
                  (ii) subject to the terms and conditions of this Agreement.
                </p>
                <h2 className="font-semibold">Maintenance and Support</h2>
                <p>
                  The Company does not provide any maintenance or support for
                  the download and use of the Application. To the extent that
                  any maintenance or support is required by applicable law, the
                  Company, not the Application Store, shall be obligated to
                  furnish any such maintenance or support.
                </p>
                <h2 className="font-semibold">Third-Party Services</h2>
                <p>
                  The Application may display, include or make available
                  third-party content (including data, information, applications
                  and other products services) or provide links to third-party
                  websites or services. You acknowledge and agree that the
                  Company shall not be responsible for any Third-party Services,
                  including their accuracy, completeness, timeliness, validity,
                  copyright compliance, legality, decency, quality or any other
                  aspect thereof. The Company does not assume and shall not have
                  any liability or responsibility to You or any other person or
                  entity for any Third-party Services. You must comply with
                  applicable Third parties' Terms of agreement when using the
                  Application. You access and use Third-party Services entirely
                  at your own risk and subject to such third parties' terms and
                  conditions. The user may incur charges from third parties by
                  using the Application. One example are network fees required
                  to use an asset network applicable to an asset transaction.
                  Alby may attempt to calculate such a fee for you. Our
                  calculation may not be sufficient, or it may be excessive. You
                  are solely responsible for selecting and paying any such fee
                  and Alby shall not advance or fund such a fee on the user's
                  behalf. Alby shall not be responsible for any excess or
                  insufficient fee calculation.
                </p>
                <h2 className="font-semibold">Term and Termination</h2>
                <p>
                  This Agreement shall remain in effect until terminated by You
                  or the Company. The Company may, in its sole discretion, at
                  any time and for any or no reason, suspend or terminate this
                  Agreement with or without prior notice. This Agreement will
                  terminate immediately, without prior notice from the Company,
                  in the event that you fail to comply with any provision of
                  this Agreement. You may also terminate this Agreement by
                  deleting the Application and all copies thereof from your
                  Device. Upon termination of this Agreement, You shall cease
                  all use of the Application and delete all copies of the
                  Application from your Device. Termination of this Agreement
                  will not limit any of the Company's rights or remedies at law
                  or in equity in case of breach by You (during the term of this
                  Agreement) of any of your obligations under the present
                  Agreement.
                </p>
                <h2 className="font-semibold">Acceptable Use</h2>
                <p>
                  The user agrees not to use the Application in ways that:
                  <ul className="list-disc list-inside">
                    <li>
                      violate, misappropriate, or infringe the rights of any
                      Alby entity, our users, or others, including privacy,
                      publicity, intellectual property, or other proprietary
                      rights;
                    </li>
                    <li>
                      are illegal, defamatory, threatening, intimidating, or
                      harassing;
                    </li>
                    <li>involve impersonating someone;</li>
                    <li>
                      breach any duty toward or rights of any person or entity,
                      including rights of publicity, privacy, or trademark;
                    </li>
                    <li>
                      involve sending illegal or impermissible communications
                      such as bulk messaging, auto-messaging, auto-dialing, and
                      the like;
                    </li>
                    <li>
                      avoid, bypass, remove, deactivate, impair, descramble or
                      otherwise circumvent any technological measure implemented
                      by us or any of our service providers or any other third
                      party (including another user) to protect the Application;
                    </li>
                    <li>
                      interfere with, or attempt to interfere with, the access
                      of any user, host or network, including, without
                      limitation, sending a virus, overloading, flooding,
                      spamming, or mail-bombing;
                    </li>
                    <li>violate any applicable law or regulation; or</li>
                    <li>
                      encourage or enable any other individual to do any of the
                      foregoing.
                    </li>
                  </ul>
                </p>
                <h2 className="font-semibold">Other Alby Services</h2>
                <p>
                  Use of the Application may require other services from Alby
                  such as an Alby Account with additional terms and fees. By
                  using this software in connection with another Alby service,
                  you agree to the applicable terms of that service, which you
                  may access and review at https://getalby.com/terms-of-service.
                  If you do not agree to the applicable terms and conditions for
                  such a service, do not use the Application in connection with
                  that service.
                </p>
                <h2 className="font-semibold">Personal Information</h2>
                <p>
                  Your submission of personal information through the
                  Application is governed by our Privacy Policy at
                  https://getalby.com/privacy-policy.
                </p>

                <h2 className="font-semibold">Indemnification</h2>
                <p>
                  You agree to indemnify and hold the Company and its parents,
                  subsidiaries, affiliates, officers, employees, agents,
                  partners and licensors (if any) harmless from any claim or
                  demand, including reasonable attorneys' fees, due to or
                  arising out of your: (a) use of the Application; (b) violation
                  of this Agreement or any law or regulation; or (c) violation
                  of any right of a third party.
                </p>
                <h2 className="font-semibold">Warranties</h2>
                <p>
                  THE APPLICATION IS PROVIDED ON AN AS-IS AND AS-AVAILABLE
                  BASIS. YOU AGREE THAT YOUR USE OF THE APPLICATION WILL BE AT
                  YOUR SOLE RISK. TO THE FULLEST EXTENT PERMITTED BY LAW, WE
                  DISCLAIM ALL WARRANTIES, EXPRESS OR IMPLIED, IN CONNECTION
                  WITH THE APPLICATION AND YOUR USE THEREOF, INCLUDING, WITHOUT
                  LIMITATION, THE IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS
                  FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT. WE MAKE NO
                  WARRANTIES OR REPRESENTATIONS ABOUT THE ACCURACY OR
                  COMPLETENESS OF APPLICATION'S CONTENT OR THE CONTENT OF ANY
                  WEBSITES LINKED TO THE APPLICATION AND WE WILL ASSUME NO
                  LIABILITY OR RESPONSIBILITY FOR ANY (1) ERRORS, MISTAKES, OR
                  INACCURACIES OF CONTENT AND MATERIALS, (2) PERSONAL INJURY OR
                  PROPERTY DAMAGE, OF ANY NATURE WHATSOEVER, RESULTING FROM YOUR
                  ACCESS TO AND USE OF THE APPLICATION, (3) ANY UNAUTHORIZED
                  ACCESS TO OR USE OF OUR SECURE SERVERS AND ANY AND ALL
                  PERSONAL INFORMATION OR FINANCIAL INFORMATION STORED THEREIN,
                  (4) ANY INTERRUPTION OR CESSATION OF TRANSMISSION TO OR FROM
                  THE APPLICATION, (5) ANY BUGS, VIRUSES, TROJAN HORSES, OR THE
                  LIKE THAT MAY BE TRANSMITTED TO OR THROUGH THE APPLICATION BY
                  ANY THIRD PARTY, OR (6) ANY ERRORS OR OMISSIONS IN ANY CONTENT
                  AND MATERIALS OR FOR ANY LOSS OR DAMAGE OF ANY KIND INCURRED
                  AS A RESULT OF THE USE OF ANY CONTENT POSTED, TRANSMITTED, OR
                  OTHERWISE MADE AVAILABLE VIA THE APPLICATION. WE WILL NOT BE A
                  PARTY TO OR IN ANY WAY BE RESPONSIBLE FOR MONITORING ANY
                  TRANSACTION BETWEEN YOU AND ANY THIRD-PARTY PROVIDERS OF
                  PRODUCTS OR SERVICES. ALBY MAKES NO GUARANTEE THAT YOUR USE OF
                  THE APPLICATION WILL COMPLY WITH ANY APPLICABLE LAW OR
                  REGULATION. Some jurisdictions do not allow the exclusion of
                  certain types of warranties or limitations on applicable
                  statutory rights of a consumer, so some or all of the above
                  exclusions and limitations may not apply to You. But in such a
                  case the exclusions and limitations set forth in this section
                  shall be applied to the greatest extent enforceable under
                  applicable law. To the extent any warranty exists under law
                  that cannot be disclaimed, the Company, not the Application
                  Store, shall be solely responsible for such warranty.
                </p>
                <h2 className="font-semibold">Limitations of Liability</h2>
                <p>
                  Alby shall not be liable to you or anyone else for any loss or
                  injury resulting directly or indirectly from your use of the
                  Application, including any loss caused in whole or part by any
                  inaccuracies or incompleteness, delays, interruptions, errors
                  or omissions, including, but not limited to, those arising
                  from the negligence of Alby Inc. or contingencies beyond its
                  control in procuring, compiling, interpreting, computing,
                  reporting, or delivering the Application. In no event will
                  Alby be liable to you or anyone else for any decision made or
                  action taken by you in reliance on, or in connection with your
                  use of the Application or the information therein. IN NO EVENT
                  WILL WE OR OUR DIRECTORS, EMPLOYEES, OR AGENTS BE LIABLE TO
                  YOU OR ANY THIRD PARTY FOR ANY DIRECT, INDIRECT,
                  CONSEQUENTIAL, EXEMPLARY, INCIDENTAL, SPECIAL, OR PUNITIVE
                  DAMAGES, INCLUDING LOST PROFIT, LOST REVENUE, LOSS OF DATA, OR
                  OTHER DAMAGES ARISING FROM YOUR USE OF THE APPLICATION, EVEN
                  IF WE HAVE BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.
                  NOTWITHSTANDING ANYTHING TO THE CONTRARY CONTAINED HEREIN, OUR
                  LIABILITY TO YOU FOR ANY CAUSE WHATSOEVER AND REGARDLESS OF
                  THE FORM OF THE ACTION, WILL AT ALL TIMES BE LIMITED TO THE
                  AMOUNT PAID, IF ANY, BY YOU TO US DURING THE SIX (6) MONTH
                  PERIOD PRIOR TO ANY CAUSE OF ACTION ARISING.{" "}
                </p>

                <p>
                  We are not responsible for loss of assets due to customer
                  error. These include, but are not limited to, loss of recovery
                  phrase, loss of credentials, configuring the same Application
                  on another device or recovering assets with backups.
                  Furthermore, we are not responsible for loss of funds due to
                  system outages. It is up to you to ensure proper backups and
                  monitoring of the Application.{" "}
                </p>
                <p>
                  {" "}
                  You expressly understand and agree that the Application Store,
                  its subsidiaries and affiliates, and its licensors shall not
                  be liable to You under any theory of liability for any direct,
                  indirect, incidental, special consequential or exemplary
                  damages that may be incurred by You, including any loss of
                  data, whether or not the Application Store or its
                  representatives have been advised of or should have been aware
                  of the possibility of any such losses arising.
                </p>

                <h2 className="font-semibold">Severability and Waiver</h2>
                <p>
                  {" "}
                  If any provision of this Agreement is held to be unenforceable
                  or invalid, such provision will be changed and interpreted to
                  accomplish the objectives of such provision to the greatest
                  extent possible under applicable law and the remaining
                  provisions will continue in full force and effect.
                </p>
                <p>
                  Except as provided herein, the failure to exercise a right or
                  to require performance of an obligation under this Agreement
                  shall not effect a party's ability to exercise such right or
                  require such performance at any time thereafter nor shall the
                  waiver of a breach constitute a waiver of any subsequent
                  breach.
                </p>

                <h2 className="font-semibold">Product Claims</h2>
                <p>
                  The Company does not make any warranties concerning the
                  Application. To the extent You have any claim arising from or
                  relating to your use of the Application, the Company, not the
                  Application Store, is responsible for addressing any such
                  claims, which may include, but not limited to: (i) any product
                  liability claims; (ii) any claim that the Application fails to
                  conform to any applicable legal or regulatory requirement; and
                  (iii) any claim arising under consumer protection, or similar
                  legislation.
                </p>
                <h2 className="font-semibold">
                  United States Legal Compliance
                </h2>
                <p>
                  You represent and warrant that (i) You are not located in a
                  country that is subject to the United States government
                  embargo, or that has been designated by the United States
                  government as a "terrorist supporting" country, and (ii) You
                  are not listed on any United States government list of
                  prohibited or restricted parties.
                </p>
                <h2 className="font-semibold">Changes to this Agreement</h2>
                <p>
                  The Company reserves the right, at its sole discretion, to
                  modify or replace this Agreement at any time. If a revision is
                  material we will provide at least 30 days' notice prior to any
                  new terms taking effect. What constitutes a material change
                  will be determined at the sole discretion of the Company. By
                  continuing to use the Application after any revisions become
                  effective, You agree to be bound by the revised terms. If You
                  do not agree to the new terms, You are no longer authorized to
                  use the Application.
                </p>

                <h2 className="font-semibold">Governing Law</h2>
                <p>
                  The laws of Switzerland, excluding its conflicts of law rules,
                  shall govern this Agreement and your use of the Application.
                  Your use of the Application may also be subject to other
                  local, state, national, or international laws.
                </p>

                <h2 className="font-semibold">Entire Agreement</h2>
                <p>
                  The Agreement constitutes the entire agreement between You and
                  the Company regarding your use of the Application and
                  supersedes all prior and contemporaneous written or oral
                  agreements between You and the Company. You may be subject to
                  additional terms of services that apply when You use or
                  purchase other Company's services.
                </p>
                <h2 className="font-semibold">Contact information</h2>
                <p>
                  Questions about the Agreement should be sent to us at
                  hello@getalby.com.
                </p>
              </div>
              <DialogFooter>
                <DialogClose asChild>
                  <Button>Close</Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>{" "}
          and{" "}
          <Dialog>
            <DialogTrigger asChild>
              <a className="underline cursor-pointer">Privacy Policy</a>
            </DialogTrigger>
            <DialogContent
              className={"lg:max-w-screen-lg overflow-y-scroll max-h-[90%]"}
            >
              <DialogHeader>
                <DialogTitle>Privacy Policy</DialogTitle>
              </DialogHeader>
              <div className="flex flex-col gap-2 text-sm">
                <p>
                  getalby.com (hereinafter "Alby" or "We" or "Us") welcomes you
                  to our internet page and services (together also referred to
                  as "Online Offers"). We thank you for your interest in our
                  company and our products.
                </p>
                <h2 className="font-semibold">1. Alby respects your privacy</h2>
                <p>
                  The protection of your privacy throughout the course of
                  processing personal data as well as the security of all
                  business data are important concerns to us. We process
                  personal data that was gathered during your visit of our
                  Online Offers confidentially and only in accordance with
                  statutory regulations. Data protection and information
                  security are included in our corporate policy.
                </p>
                <h2 className="font-semibold">2. Controller</h2>
                <p>
                  Alby is the controller responsible for the processing of your
                  data; exceptions are outlined in this data protection notice.
                  Our contact details are as follows: Alby Inc. 8 The Green STE
                  A, Dover, DE 19901, United States. Contact email:
                  hello@getalby.com
                </p>

                <h2 className="font-semibold">
                  3. Collection, processing and usage of personal data
                </h2>
                <p>
                  Communication data, transaction data, lightning node
                  authentication data are processed
                </p>
                <p>
                  3.1 Processed categories of data Communication data,
                  transaction data, lightning node authentication data are
                  processed
                </p>
                <p>
                  3.2 Principles Personal data consists of all information
                  related to an identified or identifiable natural person, this
                  includes, e.g. names, addresses, phone numbers, email
                  addresses, contractual master data, contract accounting and
                  payment data, which is an expression of a person's identity.
                  We collect, process and use personal data (including IP
                  addresses) only when there is either a statutory legal basis
                  to do so or if you have given your consent to the processing
                  or use of personal data concerning this matter, e.g. by means
                  of registration.
                </p>
                <p>
                  3.3. Processing purposes and legal basis We as well as the
                  service providers commissioned by us; process your personal
                  data for the following processing purposes:
                </p>
                <p>
                  3.3.1 Provision of these Online Offers Legal basis: Legitimate
                  interest as long as this occurs in accordance with data
                  protection and competition law. Fulfillment of contractual
                  obligations according to the Terms of Service
                </p>
                <p>
                  3.3.2 Resolving service disruptions as well as for security
                  reasons. Legal basis: Fulfillment of our legal obligations
                  within the scope of data security and legitimate interest in
                  resolving service disruptions as well as in the protection of
                  our offers.
                </p>
                <p>
                  3.3.3 Self-promotion and promotion by others as well as market
                  research and reach analysis done within the scope statutorily
                  permitted or based on consent. Legal basis: Consent or
                  legitimate interest on our part in direct marketing if in
                  accordance with data protection and competition law.
                </p>
                <p>
                  3.3.4 Product or customer surveys performed via email and/or
                  telephone subject to your prior express consent. Legal basis:
                  Consent.
                </p>
                <p>
                  3.3.5 Sending an email or SMS/MMS newsletter subject to the
                  recipient's consent Legal basis: Consent.
                </p>
                <p>
                  3.3.6 Safeguarding and defending our rights. Legal basis:
                  Legitimate interest on our part for safeguarding and defending
                  our rights.
                </p>
                <p>
                  3.4 Registration If you wish to use or get access to benefits
                  requiring to enter into the fulfillment of a contract, we
                  re-quest your registration. With your registration we collect
                  personal data necessary for entering into the fulfillment of
                  the contract (e.g. email address) as well as further data, if
                  applicable.
                </p>
                <p>
                  3.5 Log files Each time you use the internet, your browser is
                  transmitting certain information which we store in so-called
                  log files. We store log files to determine service disruptions
                  and for security reasons (e.g., to investigate attack
                  attempts) for a period of 60 days and delete them afterwards.
                  Log files which need to be maintained for evidence purposes
                  are excluded from deletion until the respective incident is
                  resolved and may, on a case-by-case basis, be passed on to
                  investigating authorities. Log files are also used for
                  analysis purposes (without the IP address or without the
                  complete IP address) see module "Advertisements and/or market
                  research (including web analysis, no customer surveys)". In
                  log files, the following information is saved:
                </p>
                <ul className="list-disc list-inside">
                  <li>
                    IP address (internet protocol address) of the terminal
                    device used to access the Online Offer;
                  </li>
                  <li>
                    Internet address of the website from which the Online Offer
                    is accessed (so-called URL of origin or referrer URL);
                  </li>
                  <li>
                    Name of the service provider which was used to access the
                    Online Offer;
                  </li>
                  <li>Name of the files or information accessed;</li>
                  <li>
                    Date and time as well as duration of recalling the data;
                  </li>
                  <li>Amount of data transferred;</li>
                  <li>
                    Operating system and information on the internet browser
                    used, including add-ons installed (e.g., Flash Player);
                  </li>
                  <li>
                    http status code (e.g., "Request successful" or "File
                    requested not found").
                  </li>
                </ul>
                <p>
                  3.6 Children This Online Offer is not meant for children under
                  16 years of age.
                </p>
                <p>3.7 Data transfer</p>
                <p>
                  3.7.1 Data transfer to other controllers Principally, your
                  personal data is forwarded to other controllers only if
                  required for the fulfillment of a contractual obligation, or
                  if we ourselves, or a third party, have a legitimate interest
                  in the data transfer, or if you have given your consent.
                  Particulars on the legal basis and the recipients or
                  categories of recipients can be found in the Section -
                  Processing purposes and legal basis. Additionally, data may be
                  transferred to other controllers when we are obliged to do so
                  due to statutory regulations or enforceable administrative or
                  judicial orders.
                </p>
                <p>
                  3.7.2 Service providers (general) We involve external service
                  providers with tasks such as sales and marketing services,
                  contract management, payment handling, programming, data
                  hosting. We have chosen those service providers carefully and
                  monitor them on a regular basis, especially regarding their
                  diligent handling of and protection of the data that they
                  store. All service providers are obliged to maintain
                  confidentiality and to comply with the statutory provisions.
                </p>
                <p>
                  3.7.3 Transfer to recipients outside the EEA We might transfer
                  personal data to recipients located outside the EEA into
                  so-called third countries. In such cases, prior to the
                  transfer we ensure that either the data recipient provides an
                  appropriate level of data protection or that you have
                  consented to the transfer. You are entitled to receive an
                  overview of third country recipients and a copy of the
                  specifically agreed-provisions securing an appropriate level
                  of data protection. For this purpose, please use the
                  statements made in the Contact section.
                </p>
                <p>
                  3.8 Duration of storage, retention periods Principally, we
                  store your data for as long as it is necessary to render our
                  Online Offers and connected services or for as long as we have
                  a legitimate interest in storing the data. In all other cases
                  we delete your personal data with the exception of data we are
                  obliged to store for the fulfillment of legal obligations
                  (e.g. due to retention periods under the tax and commercial
                  codes we are obliged to have documents such as contracts and
                  invoices available for a certain period of time).
                </p>

                <h2 className="font-semibold"> 4. Usage of Cookies</h2>
                <p>
                  In the context of our online service, cookies and tracking
                  mechanisms may be used. Cookies are small text files that may
                  be stored on your device when visiting our online service.
                  Tracking is possible using different technologies. In
                  particular, we process information using pixel technology
                  and/or during log file analysis.
                </p>
                <p>
                  4.1 Categories We distinguish between cookies that are
                  mandatorily required for the technical functions of the online
                  service and such cookies and tracking mechanisms that are not
                  mandatorily required for the technical function of the online
                  service. It is generally possible to use the online service
                  without any cookies that serve non-technical purposes.
                </p>
                <p>
                  4.1.1 Technically required cookies By technically required
                  cookies we mean cookies without those the technical provision
                  of the online service cannot be ensured. Such cookies will be
                  deleted when you leave the website.
                </p>
                <p>
                  4.1.2 Cookies and tracking mechanisms that are technically not
                  required We only use cookies and tracking mechanisms if you
                  have given us your prior consent in each case. With the
                  exception of the cookie that saves the current status of your
                  privacy settings (selection cookie). This cookie is set based
                  on legitimate interest. We distinguish between two
                  sub-categories with regard to these cookies and tracking
                  mechanisms:
                </p>
                <p>
                  4.1.3 Comfort cookies These cookies facilitate operation and
                  thus allow you to browse our online service more comfortably;
                  e.g. your language settings may be included in these cookies.
                </p>
                <p>
                  4.2 Management of cookies and tracking mechanisms You can
                  manage your cookie and tracking mechanism settings in the
                  browser. Note: The settings you have made refer only to the
                  browser used in each case.
                </p>
                <p>
                  4.2.1 Deactivation of all cookies If you wish to deactivate
                  all cookies, please deactivate cookies in your browser
                  settings. Please note that this may affect the functionality
                  of the website.
                </p>
                <h2 className="font-semibold">
                  {" "}
                  5. Data processing by App Store operators
                </h2>
                <p>
                  We do not collect data, and it is beyond our responsibility,
                  when data, such as username, email address and individual
                  device identifier are transferred to an app store (e.g.,
                  Google Web Store, Firefox Add-ons) when downloading the
                  respective application. We are unable to influence this data
                  collection and further processing by the App Store as
                  controller.
                </p>

                <h2 className="font-semibold">
                  6. Communication tools on social media platforms
                </h2>
                <p>
                  We use on our social media platform (e.g. twitter)
                  communication tools to process your messages sent via this
                  social media platform and to offer you support. When sending a
                  message via our social media platform the message is processed
                  to handle your query (and if necessary additional data, which
                  we receive from the social media provider in connection with
                  this message as your name or files). In addition we can
                  analyze these data in an aggregated and anonymized form in
                  order to better understand how our social media platform is
                  used. The legal basis for the processing of your data is our
                  legitimate interest (Art. 6 para. 1 s. 1 lit. f GDPR) or, if
                  applicable, an existing contractual relationship (Art. 6 para.
                  1 s. 1 lit. b GDPR).
                </p>
                <h2 className="font-semibold">
                  7. Newsletter with opt-in; Right of withdrawal
                </h2>
                <p>
                  Within the scope of our Online Offers you can sign up for
                  newsletters. We provide the so-called double opt-in option
                  which means that we will only send you a newsletter via email,
                  mobile messenger (such as, e.g. WhatsApp), SMS or push
                  notification after you have explicitly confirmed the
                  activation of the newsletter service to us by clicking on the
                  link in a notification. In case you wish to no longer receive
                  newsletters, you can terminate the subscription at any time by
                  withdrawing your consent. You can withdraw your consent to
                  email newsletters by clicking on the link which is sent in the
                  respective newsletter mail, or in the administrative settings
                  of the online offer. Alternatively, please contact us via the
                  contact details provided in the Contact section.
                </p>
                <h2 className="font-semibold">8. External links</h2>
                <p>
                  Our Online Offers may contain links to internet pages of third
                  parties, in particular providers who are not related to us.
                  Upon clicking on the link, we have no influence on the
                  collecting, processing and use of personal data possibly
                  transmitted by clicking on the link to the third party (such
                  as the IP address or the URL of the site on which the link is
                  located) as the conduct of third parties is naturally beyond
                  our control. We do not assume responsibility for the
                  processing of personal data by third parties.
                </p>
                <h2 className="font-semibold">9. Security</h2>
                <p>
                  Our employees and the companies providing services on our
                  behalf, are obliged to confidentiality and to compliance with
                  the applicable data protection laws. We take all necessary
                  technical and organizational measures to ensure an appropriate
                  level of security and to protect your data that are
                  administrated by us especially from the risks of unintended or
                  unlawful destruction, manipulation, loss, change or
                  unauthorized disclosure or unauthorized access. Our security
                  measures are, pursuant to technological progress, constantly
                  being improved.
                </p>
                <h2 className="font-semibold">10. User rights</h2>
                <p>
                  To enforce your rights, please use the details provided in the
                  Contact section. In doing so, please ensure that an
                  unambiguous identification of your person is possible.
                </p>
                <p>
                  10.1 Right to information and access You have the right to
                  obtain confirmation from us about whether or not your personal
                  data is being processed, and, if this is the case, access to
                  your personal data.
                </p>
                <p>
                  10.2 Right to correction and deletion You have the right to
                  obtain the rectification of inaccurate personal data. As far
                  as statutory requirements are fulfilled, you have the right to
                  obtain the completion or deletion of your data. This does not
                  apply to data which is necessary for billing or accounting
                  purposes or which is subject to a statutory retention period.
                  If access to such data is not required, however, its
                  processing is restricted (see the following).
                </p>
                <p>
                  10.3 Restriction of processing As far as statutory
                  requirements are fulfilled you have the right to demand for
                  restriction of the processing of your data.
                </p>
                <p>
                  10.4 Data portability As far as statutory requirements are
                  fulfilled you may request to receive data that you have
                  provided to us in a structured, commonly used and
                  machine-readable format or - if technically feasible - that we
                  transfer those data to a third party.
                </p>
                <p>10.5 Right of objection</p>
                <p>
                  10.5.1 Objection to direct marketing Additionally, you may
                  object to the processing of your personal data for direct
                  marketing purposes at any time. Please take into account that
                  due to organizational reasons, there might be an overlap
                  between your objection and the usage of your data within the
                  scope of a campaign which is already running.
                </p>
                <p>
                  10.5.2 Objection to data processing based on the legal basis
                  of "legitimate interest" In addition, you have the right to
                  object to the processing of your personal data at any time,
                  insofar as this is based on "legitimate interest". We will
                  then terminate the processing of your data, unless we
                  demonstrate compelling legitimate grounds according to legal
                  requirements which override your rights.
                </p>
                <p>
                  10.6 Withdrawal of consent In case you consented to the
                  processing of your data, you have the right to revoke this
                  consent at any time with effect for the future. The lawfulness
                  of data processing prior to your withdrawal remains unchanged.
                </p>
                <p>
                  10.7 Right to lodge complaint with supervisory authority: You
                  have the right to lodge a complaint with a supervisory
                  authority. You can appeal to the supervisory authority which
                  is responsible for your place of residence or your state of
                  residency.
                </p>

                <h2 className="font-semibold">
                  11. Changes to the Data Protection Notice
                </h2>
                <p>
                  We reserve the right to change our security and data
                  protection measures. In such cases, we will amend our data
                  protection notice accordingly. Please, therefore, notice the
                  current version of our data protection notice, as this is
                  subject to changes.
                </p>
                <h2 className="font-semibold">12. Contact</h2>
                <p>
                  If you wish to contact us, please find us at the address
                  stated in the "Controller" section. Effective date: 2023.03.01
                </p>
              </div>
              <DialogFooter>
                <DialogClose asChild>
                  <Button>Close</Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </div>
      </div>
    </Container>
  );
}
