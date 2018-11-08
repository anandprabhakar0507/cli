package v6_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/v6"
	"code.cloudfoundry.org/cli/command/v6/v6fakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = FDescribe("Domains Command", func() {
	var (
		cmd             DomainsCommand
		testUI          *ui.UI
		fakeConfig      *commandfakes.FakeConfig
		fakeSharedActor *commandfakes.FakeSharedActor
		fakeActor       *v6fakes.FakeDomainsActor
		// fakeActorV3     *v6fakes.FakeOrgActorV3
		binaryName string
		executeErr error
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeSharedActor = new(commandfakes.FakeSharedActor)
		fakeActor = new(v6fakes.FakeDomainsActor)
		// fakeActorV3 = new(v6fakes.FakeOrgActorV3)

		cmd = DomainsCommand{
			UI:          testUI,
			Config:      fakeConfig,
			SharedActor: fakeSharedActor,
			Actor:       fakeActor,
			// ActorV3:     fakeActorV3,
		}

		binaryName = "some-binary-name"
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	When("the user is not logged in", func() {
		BeforeEach(func() {
			fakeSharedActor.CheckTargetReturns(actionerror.NotLoggedInError{BinaryName: binaryName})
		})

		It("should show failure message and return an error", func() {
			Expect(executeErr).To(MatchError(actionerror.NotLoggedInError{BinaryName: binaryName}))
			Expect(fakeSharedActor.CheckTargetCallCount()).To(Equal(1))

			checkTargetedOrgArg, checkTargetedSpaceArg := fakeSharedActor.CheckTargetArgsForCall(0)

			Expect(checkTargetedOrgArg).To(BeTrue())
			Expect(checkTargetedSpaceArg).To(BeFalse())
		})
	})

	When("the user is logged in and targeted an Org", func() {
		When("getting the current user fails", func() {
			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(configv3.User{}, errors.New("get-user-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("get-user-error"))
			})
		})

		When("getting the current user succeeds", func() {
			var targetedOrg configv3.Organization

			BeforeEach(func() {
				fakeConfig.CurrentUserReturns(
					configv3.User{Name: "some-user"},
					nil)
				targetedOrg = configv3.Organization{Name: "some-org", GUID: "some-org-guid"}
				fakeConfig.TargetedOrganizationReturns(targetedOrg)
			})

			It("displays a message indicating that it is getting the domains", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Out).To(Say(`Getting domains in org some-org as some-user\.\.\.`))
			})

			When("getting the shared domains", func() {
				When("GetDomains returns an error", func() {
					BeforeEach(func() {
						fakeActor.GetDomainsReturns([]v2action.Domain{}, v2action.Warnings{"warning-1", "warning-2"}, actionerror.OrganizationNotFoundError{Name: targetedOrg.Name})
					})

					It("fails and returns an error", func() {
						Expect(executeErr).To(MatchError(actionerror.OrganizationNotFoundError{Name: targetedOrg.Name}))
						actualOrgGUID := fakeActor.GetDomainsArgsForCall(0)
						Expect(actualOrgGUID).To(Equal(targetedOrg.GUID))
						Expect(fakeActor.GetDomainsCallCount()).To(Equal(1))
					})

					It("displays all warnings", func() {
						Expect(testUI.Out).To(Say(`warning-1`))
						Expect(testUI.Out).To(Say(`warning-2`))
					})
				})

				When("GetDomains returns more than one domain", func() {
					var (
						privateDomain v2action.Domain
						sharedDomain  v2action.Domain
					)

					BeforeEach(func() {
						privateDomain = v2action.Domain{
							Name: "private.domain",
							Type: constant.PrivateDomain,
						}

						sharedDomain = v2action.Domain{
							Name: "shared.domain",
							Type: constant.SharedDomain,
						}
						fakeActor.GetDomainsReturns([]v2action.Domain{privateDomain, sharedDomain}, v2action.Warnings{"warning-1", "warning-2"}, nil)
					})

					It("displays all domains", func() {
						Fail("NYI")
					})

					It("displays all warnings", func() {
						Expect(testUI.Out).To(Say(`warning-1`))
						Expect(testUI.Out).To(Say(`warning-2`))
					})
				})
			})
		})
	})
})
