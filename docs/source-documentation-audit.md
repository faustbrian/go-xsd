# Source documentation audit

[Documentation index](README.md) · [Repository entry](../README.md)

## Ownership and prewrite acceptance

Tier A comments/current-prose only at freshly fetched main `a778100a312841fee263bb47efb8f4070639d30a`. Owner: this standalone `/v2` XSD family; root parses and serializes a mutable schema model, builder assembles it, compile owns immutable resolved sets, validate owns bounded per-call instance work, datatype owns lexical/value-space operations, resolve owns explicit capabilities, and xsdtest owns opt-in trusted fixture ingestion. WSDL/SOAP consumers depend on these public contracts; no sibling checkout or consumer execution is required for unchanged signatures/runtime.

Acceptance: audit every tracked Go file and public/private critical boundary without changing executable tokens, AST, literals, directives, XML/schema/SQL fixtures, generated outputs, API export, dependencies, source pins, normative decision history or security authority. Explain actual XML Schema 1.0 support and limits, QName/namespaces, include/import identity, mutable-model ownership, immutable-set copying, traversal/cancellation, resource preflight, regexp/numeric/calendar caveats and privacy-safe versus trusted diagnostics. No runtime repairs, parser probes, standards campaigns, services, fuzz/race/benchmark or release qualification.

Smallest selected checks: Go 1.27.2 formatting/all-source identity and comment rendering, eight native package GoDocs, the single existing output-bearing Example, actual native documentation/spelling/local links and offline specification validation, lightweight structural source checker, pinned changed-file/added/staged secrets and one ordinary complete final diff. Obtain serialized Go grant first; use and remove exact three owned caches and helper outputs.

## Initial gaps versus questions

Initial source checker finds 323 malformed/missing exported comments, concentrated in mutable model fields and compiler/validator/options/result descriptors. Existing package headers are present but often omit useful ownership and specification limits. Native API mode is `apidiff` using `api/v2.export`; it is not a full-prose snapshot and remains unchanged. Root Make exposes no separate docs Example recipe; select the one actual Example directly. No generated Go or generation directive appears in the initial inventory. Two benchmark go:embed directives are immutable fixture ownership, not producers.

Investigate before documenting: defaults versus negative limits and disabled zero controls, cloned versus borrowed models/maps/decimal coefficients, imported graph document identity and resolver returned-URI checks, symlink/path capability confinement, blocking readers and cooperative context only, standard regexp matching and Unicode table authority, exact decimal scale/canonicalization, partial writer outcomes, typed/redacted error boundaries and explicit trusted XSTS reports. Unknown rationale or unimplemented normative behaviors must remain questions, never new desired-policy/history edits. Dated security/goal/evidence and original failed CI history remain immutable.

## Reproducible complete inventory

Reproduce with `git ls-files -- "*.go"` and tracked `go.mod` inventory; cross-check `modules.json` and `packages.json`. One root module `github.com/faustbrian/go-xsd/v2` contains eight source directories: `.`, `builder`, `compile`, `datatype`, `internal/errsafe`, `resolve`, `validate`, `xsdtest`. There are no noncatalog Go directories or nested modules. Rows include assurance entry names; test sources are inspected for contracts and helper ownership, not newly qualified by a documentation pass.

| Tracked file | Role / assurance entry points | Initial gaps |
| --- | --- | --- |
| `aaa_mutation_preflight_internal_test.go` | test; TestRepresentativeSchemaParsesBeforeParallelCampaigns | 0 |
| `benchmark_test.go` | test; BenchmarkParseSchema, BenchmarkCompileSchema, BenchmarkValidateInstance, BenchmarkMarshalSchema | 0 |
| `builder/builder.go` | source | 0 |
| `builder/builder_test.go` | test; TestBuildProducesAnIsolatedCompilableSchema, TestBuilderRejectsMalformedAndDuplicateNames, TestBuilderSupportsAttributesAndFormDefaults, TestBuilderRejectsInvalidFormsAndMissingTypes, TestBuilderDeepCopiesNestedAnonymousTypes, TestBuildRejectsAnInvalidGeneratedSchema | 0 |
| `builder/clone_internal_test.go` | test; TestBuildPropagatesSerializationFailures, TestCloneSimpleTypeRecursivelyIsolatesNestedDefinitions, TestCloneSimpleTypePreservesNestedPointerCycles, TestCloneComplexTypeIsolatesAttributeGroupAnnotations, TestCloneComplexTypeIsolatesSimpleContentRestrictions, TestCloneElementIsolatesValueNamespaces, TestCloneAttributeUsesIsolatesValueNamespaces | 0 |
| `compile/accessor_internal_test.go` | test; TestDocumentAccessorsReturnIsolatedCopies, TestCollectModelIdentityConstraintsTraversesNestedContent, TestResolutionPrimitiveDecisionTables, TestIdentityXPathGrammarDecisionTables | 0 |
| `compile/boundary_internal_test.go` | test; TestRequiredCompositionNamespaces, TestWildcardDefinitionValidation, TestAttributeUseExpandedName, TestAttributeUseRejectsDirectNotation | 0 |
| `compile/cancellation_internal_test.go` | test; TestNullableModelGroupHonorsCompilationOwnerCancellation, TestCompileCancellationAfterGraphLoading, TestCompileCancellationDoesNotPublishSetOrPoisonCompiler | 0 |
| `compile/clone_internal_test.go` | test; TestCompiledComponentClonesIsolateAnnotations, TestCompiledAttributeCloneIsolatesValueNamespaces, TestCompiledGroupAccessorsDeepCopyMetadata | 0 |
| `compile/clone_owner_cancellation_test.go` | test; TestCloneOwnersHonorCancellationWithoutChangingContextFreeCopies | 0 |
| `compile/compile.go` | source | 21 |
| `compile/compile_test.go` | test; TestCompileResolvesCyclesOnceAndAppliesChameleonNamespace, TestCompileIndexesChameleonComponentsByEffectiveNamespace, TestCompiledSetExposesDeterministicComponentInventory, TestCompileRejectsDuplicateNotationComponents, TestCompileEnforcesNotationValueSpace, TestCompileRejectsDuplicateTypeComponentsAcrossKinds, TestCompiledComplexTypesDoNotExposeMutableModelStorage, TestCompiledGroupsDoNotExposeMutableModelStorage, TestCompiledSimpleTypesDoNotExposeMutableModelStorage, TestCompileNormalizesChameleonLocalNamesAndReferences, TestCompileRejectsInvalidAndUnresolvedComponentContracts, TestCompileComplexContentRestriction, TestCompileAnonymousComplexContentExtension, TestCompileRejectsDuplicateAnonymousComplexTypeAttributes, TestCompileResolvesLocationlessImportsThroughCatalog, TestCompileRejectsAmbiguousContentModels, TestCompileIncludesSubstitutionMembersInParticleAttribution, TestCompileUsesDenyResolverByDefault, TestCompileAcceptsCurrentNodeIdentitySelector, TestCompileRejectsDeclarationPropertiesOnElementReference, TestCompileRejectsInvalidElementValueConstraints, TestCompileRejectsInvalidAttributeValueConstraints, TestCompileComparesValueConstraintsInTheDeclaredValueSpace, TestCompileEnforcesSimpleTypeFinal, TestCompileAllowsEmptyRestrictionOfNullableContent, TestCompileRejectsInvalidMixedSimpleContentRestrictions, TestCompileRejectsInconsistentAttributeUseSets, TestCompileAllowsNarrowerElementTermsInComplexRestrictions, TestCompileBoundsSchemaCount | 0 |
| `compile/compiler_mutation_internal_test.go` | test; TestCompilerResourceLimitsAcceptExactBoundaryAndRejectOneBeyond, TestLoadAccumulatesBytesAtTheExactLimit, TestCompileDocumentCountersUseExactLimits, TestComponentCounterUsesExactLimit, TestAnonymousElementRejectsNamedAndInlineComplexTypesTogether, TestUPAStateTracksPositionsAndOccurrenceBoundaries, TestUPAGroupComputesNullableStatesAndAllPairs, TestHealthyOwnerAllGroupPreservesUPAPositions, TestHealthyOwnerDepthWrappersAcceptSimpleString, TestExplicitAndImplicitValueConstraintFlags, TestAnonymousComplexTypeComposesReferencedWildcards, TestAnonymousComplexExtensionComposesBaseWildcards, TestSimpleContentRestrictionRejectsWildcardWithoutBaseWildcard, TestElementAndAttributeTypeDefaults, TestAnonymousComplexElementValidatesNestedAttributeUses, TestParticleLimitAcceptsExactCount, TestElementValueConstraintsAcceptSimpleContentBases, TestCompilableReferenceKinds | 0 |
| `compile/complex_type_internal_test.go` | test; TestCompileComplexTypeRejectsInvalidDerivations, TestCompileComplexTypeCombinesExtensionWildcards | 0 |
| `compile/component_validation_internal_test.go` | test; TestValidateSimpleTypeDefinitionRejectsEveryInvalidShape, TestFacetRestrictionHelperBoundaries, TestOrderedFacetRestrictionDecisionTables, TestFacetHelperBoundaryDecisions, TestValidateIdentityConstraintsRejectsEveryInvalidShape | 0 |
| `compile/coverage_internal_test.go` | test; TestSetLookupsReportMissingComponents, TestSubstitutionValidationCoversRecursiveTransitiveAndBlockedMembers, TestCompilerAndAnonymousTypeBranches, TestCompilerCoversSetAndDerivationBranches, TestCompilerCoversValidationAndIndexingBranches, TestCompilerCoversNormalizationBranches | 0 |
| `compile/derivation_internal_test.go` | test; TestApplySimpleContentDerivation, TestSimpleContentExtensionRejectsProgrammaticFacets, TestTypeDerivationMethodChains, TestElementTypeDerivationMethodSources, TestSetElementTypeDerivationMethodSources | 0 |
| `compile/doc.go` | source | 0 |
| `compile/error_privacy_test.go` | test; TestCompilerDefaultErrorPrivacy, TestCompilerPrivacyPreservesClassification, TestPublicErrorOwnersPreserveExpiredDeadline | 0 |
| `compile/expand_groups_internal_test.go` | test; TestExpandGroupsPropagatesReferenceFailures, TestExpandGroupsCombinesAttributeWildcards | 0 |
| `compile/facet_mutation_internal_test.go` | test; TestConstraintComparisonExhaustiveBoundaries, TestCompileDepthBoundary, TestCompilerLimitDefaultsAndNegativeValues, TestDeterministicOrderingComparators, TestConstraintNumericFacetsExhaustiveBoundaries, TestConstraintAtomicEqualityExhaustiveBoundaries, TestRestrictionFacetRelationsAtEveryBoundary, TestFacetRecursionStopsAtTheExactDepthBoundary, TestQNameAndUnionEqualityRejectOneSidedValidity, TestConstraintValidityPairDistinguishesEveryState, TestIntegerFacetRestrictionDirections, TestConstraintFacetLengthsAndEnumerationBoundaries, TestOrderedFacetHelpersExhaustiveBoundaries, TestOrderedFacetValidationContinuesToLaterFixedFacets, TestProcessContentsStrengthOrdering, TestFacetApplicabilityCompleteMatrix, TestWhitespaceFacetSemantics | 0 |
| `compile/facets.go` | source | 0 |
| `compile/index_redefine_internal_test.go` | test; TestApplyRedefinitionRejectsEveryInvalidTarget, TestApplyRedefinitionPreservesNonSelfAttributeReferences, TestApplyRedefinitionValidatesEachSelfDerivationProperty, TestApplyRedefinitionInheritsWildcardsAndReplacesSelfGroupReferences, TestIndexComponentsAppliesElementAndComplexTypeDefaults, TestCompileDocumentBoundaries, TestIndexComponentsRejectsMissingNamesAndDuplicates, TestChameleonNormalizationAdoptsNestedReferences, TestChameleonIndexingAdoptsGlobalNestedReferences | 0 |
| `compile/load_internal_test.go` | test; TestLoadResourceBranches, TestLoadOwnerCancellationPreventsCacheAndResolverWork, TestLoadCancellationDuringResolutionStopsResourcePublication, TestCompileCancellationDuringResolutionDiscardsSetAndAllowsReuse | 0 |
| `compile/model_group_internal_test.go` | test; TestValidateModelGroupRejectsInvalidParticles, TestValidateModelGroupEnforcesParticleLimitRecursively, TestAllCompositorRejectsEveryInvalidOccurrenceShape | 0 |
| `compile/parser_admission_test.go` | test; TestCompilerPropagatesPerDocumentParseAdmission | 0 |
| `compile/particle_admission_internal_test.go` | test; TestExpandedParticleFinalCountAndCopyIsolation, TestFinalParticleLimitAllowsReducingRedefinition, TestParticleCopyAdmission, TestParticleCopiesCountWithoutCompiler, TestParticleCopyBudgetCumulativeAndWrappers, TestParticleCopyBudgetPublicControls | 0 |
| `compile/public_completion_test.go` | test; TestCompilePropagatesRedefinitionFailure, TestPreCanceledCompileSkipsURIAdmission | 0 |
| `compile/reachable_tail_internal_test.go` | test; TestCompilerAndSetReachableBoundaries, TestSubstitutionMemberRejectsInvalidAffiliationChains, TestCompileSubstitutionsRejectsMissingAndUnrelatedTypes, TestAnonymousComplexTypeReachableBranches, TestParticleAndAttributeHelpersCoverRemainingBranches, TestCompilerCoversRemainingAnonymousAndRecursiveBranches, TestSubstitutionMemberStopsWhenAffiliationEnds | 0 |
| `compile/restriction_internal_test.go` | test; TestValidateComplexRestrictionBranches, TestExpandAttributeGroupBranches, TestAttributeRestrictionRules, TestSimpleTypeDerivationRules, TestModelGroupRestrictionRules, TestParticleRestrictionRules, TestModelGroupNullability, TestElementTermRestrictionRules, TestWildcardSetOperationsAndOverlap, TestWildcardRestrictionAndNamespaceRules | 0 |
| `compile/uri_admission_test.go` | test; TestCompilerDefaultURIAdmission, TestCompilerURIAdmissionBoundaries, TestCompilerResolvedURIAtExactAllowance | 0 |
| `compile/validate_components_internal_test.go` | test; TestValidateComponentsPropagatesEveryComponentFailure | 0 |
| `compile/value_constraint_internal_test.go` | test; TestConstraintFacetValueDecisionTables, TestConstraintValueEqualityDecisionTables, TestInlineValueConstraintValidation, TestNamedValueConstraintValidation, TestValidateAttributeUseDecisionTable, TestAttributeValueConstraintBoundaryBranches, TestValidateElementValueConstraintDecisionTable | 0 |
| `compile/wsdl_contract_test.go` | test; TestCompiledSetSatisfiesWSDLConsumerContract | 0 |
| `components_test.go` | test; TestParseCapturesGlobalDeclarationsAndTypeDefinitions | 0 |
| `composition_test.go` | test; TestParsePreservesCompositionDirectivesAndResolvedBaseURIs | 0 |
| `datatype/calendar.go` | source | 0 |
| `datatype/coverage_internal_test.go` | test; TestOrderedWrappersCoverCalendarFormsAndNegativeDurations, TestPatternEscapesCoverComplementClasses | 0 |
| `datatype/decimal.go` | source | 10 |
| `datatype/decimal_test.go` | test; TestParseDecimalUsesExactValueAndCanonicalForm, TestParseDecimalRejectsOutsideLexicalSpace, TestDecimalComparisonDoesNotLosePrecision, TestDecimalFacetsUseValueSpaceAndCanonicalDigits, TestEveryDecimalFacetRejectsItsBoundary | 0 |
| `datatype/derivation.go` | source | 0 |
| `datatype/derivation_test.go` | test; TestBuiltInDerivationMetadata | 0 |
| `datatype/doc.go` | source | 0 |
| `datatype/fuzz_test.go` | test; FuzzParseDecimal, FuzzBuiltInLexical | 0 |
| `datatype/integer.go` | source | 1 |
| `datatype/integer_test.go` | test; TestParseIntegerCanonicalizesArbitraryPrecisionValues, TestParseIntegerRejectsDecimalAndExponentLexicalForms, TestBuiltInIntegerRangesAreExact, TestIntegerComparisonAndUnknownRange | 0 |
| `datatype/internal_test.go` | test; TestCalendarValidatorRejectsUnknownPrimitive, TestRangeTableSetExpandsStrides | 0 |
| `datatype/lexical.go` | source | 0 |
| `datatype/lexical_test.go` | test; TestValidateBuiltInLexicalSpaces, TestValidateBuiltInLexicalRejectsUnknownType, TestValidateCalendarAndDurationLexicalSpaces | 0 |
| `datatype/mutation_internal_test.go` | test; TestCalendarLexicalBoundaries, TestCalendarHelperBoundaries, TestDecimalInternalBoundaries, TestIntegerExactLexicalLimit, TestLexicalHelperBoundaries, TestOrderedDurationAndCalendarBoundaries, TestPatternTranslatorAndRuneSetBoundaries | 0 |
| `datatype/order.go` | source | 0 |
| `datatype/order_test.go` | test; TestOrderedValuePublicBoundaries | 0 |
| `datatype/pattern.go` | source | 0 |
| `datatype/pattern_admission_test.go` | test; TestCompilePatternContextCancellation, TestCompilePatternClassDepthAdmission, TestCompilePatternSiblingClassesDoNotAccumulateDepth | 0 |
| `datatype/pattern_blocks.go` | source | 0 |
| `datatype/pattern_test.go` | test; TestCompilePatternUsesXMLSchemaSemantics, TestCompilePatternBoundsTranslationWork, TestCompilePatternRejectsInvalidExpressions | 0 |
| `datatype/range_count_test.go` | test; TestBoundedRangeCount, TestRangeTableSetRejectsInvalidRanges | 0 |
| `diagnostic_privacy_test.go` | test; TestDiagnosticAndLocationDefaultPrivacy, TestDiagnosticAndLocationInvalidFormatBoundary, TestDiagnosticAndLocationNilZeroAndNestedPrivacy | 0 |
| `differential_test.go` | test; TestReferenceDifferentialCorpus | 0 |
| `doc.go` | source | 0 |
| `document_test.go` | test; TestParseErrorDefaultsPreserveTrustedInspection, TestParseErrorDefaultSurfacesDoNotEvaluateCause, TestParseErrorNilAndZeroDefaults, TestParseRejectsDTDWithoutResolvingEntities, TestParseCapturesSchemaDocumentPolicies, TestParseRejectsNonSchemaRoot, TestParseRejectsUnknownSchemaComponents, TestParseRejectsDeclarationsInsideRedefine, TestParseEnforcesSchemaCompositionGrammar, TestParseRejectsDuplicateExclusiveChildren, TestParseRejectsOutOfOrderChildren, TestParseRejectsUnknownSchemaAttributes, TestParseRejectsInvalidAndDuplicateSchemaIDs, TestParseAndMarshalPreserveAppInfo, TestParseAndMarshalPreserveDocumentationMarkup, TestMarshalRejectsUnsafeAppInfoContent, TestMarshalRejectsUnsafeDocumentationContent, TestParseRejectsMisplacedOrRepeatedComponentAnnotations, TestParseRejectsMultipleModelGroupCompositors, TestParseValidatesNotationDeclarationsAndReferences, TestParseAndMarshalPreserveNotationDeclarations, TestParseAndMarshalPreserveComponentAnnotations, TestParseRejectsUseOnGlobalAttribute, TestParseResolvesQNameAttributesInTheirNamespaceScope, TestQNameValueNamespaceScopeSurvivesSerialization | 0 |
| `errors.go` | source | 13 |
| `example_test.go` | test; Example | 0 |
| `fuzz_test.go` | test; FuzzParseSchema, FuzzValidateInstance | 0 |
| `identity_test.go` | test; TestParseCapturesIdentityConstraints | 0 |
| `identity_xpath.go` | source | 0 |
| `identity_xpath_test.go` | test; TestNormalizeIdentityXPathWhitespace | 0 |
| `internal/errsafe/error.go` | source | 0 |
| `internal/errsafe/error_test.go` | test; TestDefaultOutputRetainsOnlyCategoryAndTrustedCause | 0 |
| `leak_test.go` | test; TestCompileAndValidateDoNotLeakGoroutines | 0 |
| `model.go` | source | 224 |
| `model_test.go` | test; TestDerivationSetReportsMembershipAndLexicalValue, TestParserAndSerializerCoverNestedQNameBranches | 0 |
| `parse.go` | source | 5 |
| `parse_admission_test.go` | test; TestParseOwnedNamespaceAdmission, TestParseOwnedAnnotationAdmission, TestParseOwnedAdmissionOptionsAndContext, TestParseOwnedStringAdmission, TestParseOwnedDuplicateQNameAdmission, TestParseOwnedWorkAdmissionBoundaries, TestParseOwnedNamespaceWorkBoundaries, TestParseRootMetadataIsNotANamespaceDeclaration | 0 |
| `parse_annotation_public_test.go` | test; TestParseAppInfoMetadataNamespaces, TestParseAppInfoMetadataAdmission, TestParseIncompleteAnnotationsPreserveLocatedSyntaxCause | 0 |
| `parse_budget.go` | source | 0 |
| `parse_internal_test.go` | test; TestParseInlineSimpleTypesBranches, TestParseSimpleTypeBranches, TestParseEntryBoundaries, TestForeignExtensionElementsAreRejectedAtEverySchemaBoundary, TestNamespaceScopePreservesParentAndCapturesDeclarations, TestParsePreservesValueNamespaceScopeForDefaultAndFixedValues, TestAnnotationOpaqueContentBypassesSchemaGrammarValidation, TestSimpleContentRestrictionGrammarDoesNotLeakIntoOtherDerivations, TestParseDocumentPropagatesComponentErrors, TestParseRejectsUnknownAttributesAtGrammarBoundaries, TestParseContentDerivationBranches, TestParseRedefinitionPropagatesAnnotationErrors, TestParseComplexTypeBranches, TestParseDerivationBodyRejectsInvalidChildren, TestParseCompositorDecisionTable, TestParseFacetKindDecisionTable, TestParseRestrictionFacetBranches, TestParseModelGroupBranches, TestSetModelGroupOccurrenceBranches, TestParseRedefinitionBranches, TestSchemaReferenceResolutionBranches, TestReferenceKindDecisionTable, TestAnnotationParsingBranches, TestAnnotationChildrenAndAttributeBodyBranches, TestNotationParsingBranches, TestParserPrimitiveDecisionTables, TestParticleAttributeAndWildcardParsingBranches, TestDeclarationAttributesRejectInvalidLexicalValues, TestSchemaAttributesRejectInvalidLexicalValues, TestIdentityConstraintErrorBranches, TestNestedComponentErrorBranches, TestParserBoundaryBranches, TestParseSimpleContentRestrictionFailures, TestContextReaderChecksCancellation | 0 |
| `parse_lifecycle_internal_test.go` | test; TestParsePreparedOwnerCancellation, TestParseValidatedWithoutPreparedOwner, TestParseOwnerTokenPreservesAdmissionFailure | 0 |
| `particles_test.go` | test; TestParseCapturesNestedParticlesAndAttributeUses, TestParseCapturesTopLevelModelGroupOccurrences | 0 |
| `resolve/admission_test.go` | test; TestMemoryConstructorAdmission, TestCatalogConstructorAdmission, TestConstructorAdmissionOptions, TestConstructorAdmissionStringBytes, TestAdmittedResolversCompileOrdinaryImport | 0 |
| `resolve/catalog.go` | source | 2 |
| `resolve/chain_cancellation_test.go` | test; TestChainCancellationStopsFallbackAndAllowsReuse | 0 |
| `resolve/doc.go` | source | 0 |
| `resolve/error_privacy_test.go` | test; TestResolverPrivacyPreservesFallbackAndClassification, TestFileDefaultPrivacyPreservesTrustedOSCause, TestResolverDefaultErrorPrivacy | 0 |
| `resolve/file.go` | source | 2 |
| `resolve/file_internal_test.go` | test; TestFilePathFromURIPathNormalizesOnlyWindowsDrivePaths, TestConfinedRelativeRejectsPathAndPlatformErrors | 0 |
| `resolve/resolve.go` | source | 12 |
| `resolve/resolve_test.go` | test; TestFileResolverReadsOnlyWithinItsConfiguredRoot, TestFileResolverSupportsEscapedFileNames, TestFileResolverAcceptsAResourceAtTheExactByteLimit, TestFileResolverRejectsUnsafeAndOversizedResources, TestFileResolverRejectsSymlinkEscape, TestFileResolverReportsUnreadableResources, TestFileResolverObservesCancellationAfterReading, TestNewFileRejectsInvalidOptions, TestCatalogResolvesLocationlessImportsByNamespace, TestCatalogRejectsInvalidAndUnknownMappings, TestDenyResolverRejectsEveryResource, TestMemoryResolverReturnsOwnedResourceCopies, TestMemoryResolverHonorsCancellationAndMissingResources, TestResolversRejectInvalidConfigurationAndHonorCancellation, TestChainTriesOnlyMissingResolvers | 0 |
| `resolve/uri_admission_test.go` | test; TestFileDefaultURIAdmission, TestFileURIAdmissionBoundaries | 0 |
| `serialize.go` | source | 0 |
| `serialize_internal_test.go` | test; TestSerializerPropagatesWriteFailuresAtEveryBoundary, TestSimpleTypeSerializerPropagatesNestedFailures, TestSerializerRejectsUnknownQNameNamespaces, TestSerializerCoversProgrammaticModelChoices, TestSerializerPropagatesFragmentFlushFailures, TestMarshalPropagatesFinalFlushFailure, TestMarshalLimitInternalsFailClosed, TestMarshalLimitsTrackExactRemainingWork, TestNamespacePrefixAllocationUsesFirstAvailablePrefix, TestSerializerBuildsCanonicalNamespaceRegistry, TestSchemaSerializationDistinguishesSystemAndBaseURIs, TestRedefinitionSerializationConsumesBodiesExactlyOnce, TestRedefinitionReferenceFallbackPropagatesWriteFailure, TestIdentityConstraintSerializationHandlesSparseFieldMetadata, TestComplexTypeSerializationPreservesIndependentContentFlags, TestSerializerRejectsUnknownNamespacesAtComponentBoundaries, TestSerializerPropagatesComponentAnnotationFailures, TestSerializerPropagatesNestedComponentFailures, TestMarshalRejectsNilDocument, TestSerializerCoversRemainingModelMetadata, TestSerializerPropagatesRemainingNestedFailures | 0 |
| `serialize_limits.go` | source | 3 |
| `serialize_test.go` | test; TestMarshalWithOptionsEnforcesResourceLimits, TestMarshalWithOptionsRejectsInvalidLimits, TestMarshalIsDeterministicAndRoundTripsSchemaComponents, TestMarshalAllocatesPrefixesForProgrammaticQNames, TestMarshalPreservesEmptyValueConstraints, TestMarshalPreservesComponentIDs, TestMarshalPreservesAnnotationContentIDs, TestMarshalPreservesSimpleContentRestrictionTypesAndFacets, TestMarshalPreservesRedefinitionBodies, TestMarshalRoundTripsEveryComponentFamily | 0 |
| `simple_types_test.go` | test; TestParseCapturesRestrictionFacetsListsAndUnions | 0 |
| `validate/admission_contract_test.go` | test; TestValidateTreeOwnedPayloadAdmission, TestValidateTreePayloadOccurrences, TestValidateNamespaceAdmissionParity, TestValidateNamespaceRebindingWork, TestValidateReaderOwnedPayloadAndCancellation, TestValidateTreeExplicitQNameScope, TestValidateReaderTextPrefixWork, TestValidateTreeChildAdmissionBeforeTraversal | 0 |
| `validate/attributes_internal_test.go` | test; TestValidateAttributesStopsWhenDiagnosticsLimitIsReached, TestUntypedAttributeFixedValueEquality | 0 |
| `validate/boundary_internal_test.go` | test; TestNewRejectsMissingSetAndNegativeLimits, TestParseInstanceEnforcesEveryResourceBoundary | 0 |
| `validate/cancellation_internal_test.go` | test; TestValidationOwnerStopsCanceledWork | 0 |
| `validate/content_internal_test.go` | test; TestValidateElementPropagatesEveryContentFailure, TestValidateComplexPropagatesEveryNestedFailure | 0 |
| `validate/coverage_internal_test.go` | test; TestValidatorCoversElementTypeAndNilBranches, TestValidatorCoversSimpleIdentityAndFacetBranches, TestValidatorCoversWildcardAndNullableParticleBranches | 0 |
| `validate/derived_types_regression_test.go` | test; TestValidateNamedTypeSemantics | 0 |
| `validate/diagnostic_privacy_test.go` | test; TestValidateGeneratedDiagnosticDefaultPrivacy | 0 |
| `validate/doc.go` | source | 0 |
| `validate/identity_internal_test.go` | test; TestValidateIdentityConstraintsStopsAtEveryLimitBoundary, TestValidateIdentityConstraintsStopsAtEveryDiagnosticBoundary, TestValidateIdentityConstraintsContinuesAfterMultipleKeyrefField | 0 |
| `validate/model_matching_internal_test.go` | test; TestModelMatchingPropagatesNestedValidationFailures, TestModelMatchingCoversUnmatchedTerms | 0 |
| `validate/mutation_internal_test.go` | test; TestOrderedLimitComparisonsRespectExactBoundary, TestBoundedIdentityArithmeticRejectsOverflowAndPreservesBoundary, TestConsumeIdentityWorkUsesBoundedArithmetic, TestValidityClassificationAndConjunctionTruthTables, TestComparisonFacetBoundaries, TestOccurrenceAndIndexBoundaries, TestWildcardNamespaceConstraints, TestSortedAttributeNamesOrdersNamespaceThenLocalName, TestMutationBoundaryHelpers, TestCanonicalBinaryAndIdentityNamespaceBoundaries, TestValidationCoversRemainingSchemaBoundaries, TestNumericFacetValidationRejectsMalformedBounds, TestStrictWildcardAcceptsValidAnonymousAttribute, TestSimpleValueEqualityDistinguishesEveryPrimitiveBoundary, TestFacetLengthsUseValueSpaceAndExactBounds, TestDurationComponentsApplySignToMonthsAndSeconds, TestMissingAttributeReferenceDoesNotStopLaterUseValidation | 0 |
| `validate/pattern_context_test.go` | test; TestPatternOwnersPreserveContextAndOrdinaryAssessment | 0 |
| `validate/populated_cancellation_internal_test.go` | test; TestPopulatedFacetsStopAfterCancellation, TestPopulatedIdentityStateStopsAfterCancellation, TestCompletedTreeCloneRejectsCanceledAssessment | 0 |
| `validate/reachable_tail_internal_test.go` | test; TestValidatorCoversRemainingReachableBranches, TestValidateRootPropagatesIDReferenceLimit, TestNilledNamedComplexTypeValidatesAttributes | 0 |
| `validate/validate.go` | source | 19 |
| `validate/validate_test.go` | test; TestValidateReaderCancellationAtEOF, TestValidateReaderCancellationAfterParsedPrefix, TestValidateReaderMatchesByteValidation, TestValidateReaderEnforcesInputBoundaries, TestValidateChecksGlobalElementDatatypeExactly, TestValidateAcceptsDerivedBuiltInXSIType, TestValidateTreeMatchesByteValidation, TestValidateTreeEnforcesOwnedInputBounds, TestValidateSimpleContentRejectsUndeclaredAttributes, TestValidateRejectsDuplicateExpandedAttributeNames, TestValidateRejectsUnknownAndAbstractRoots, TestValidateRejectsDTDBeforeInstanceProcessing, TestValidateEmptyComplexTypeRejectsChildContent, TestValidateComplexSequenceChoiceAndRequiredAttributes, TestValidateAllCompositorAcceptsAnyOrderAndRequiresMembers, TestValidateChoiceAppliesAlternativeOccurrenceConstraints, TestValidateXSITypeDerivationAndBlock, TestValidateNillableDefaultAndFixedElementConstraints, TestValidateEmptyAndMixedFixedElementConstraints, TestValidateFixedConstraintsUseAnonymousValueSpaces, TestValidateFixedConstraintsUseBuiltInValueSpaces, TestValidatePatternUsesXMLSchemaMatchingSemantics, TestValidateAppliesDerivedWhitespaceBeforeOtherFacets, TestValidateCombinesPatternsByDerivationStep, TestValidateAttributeFixedValueAndGlobalReference, TestValidateRestrictionListAndUnionSimpleTypes, TestValidateQNameValuesRequireBoundPrefixes, TestValidateBuiltInLexicalSpaces, TestValidateComplexContentExtensionIncludesBaseContent, TestValidateExpandsModelAndAttributeGroups, TestValidateElementAndAttributeWildcards, TestValidateKeyUniqueAndKeyrefConstraints, TestValidateKeyrefUsesPropagatedDescendantKeyTables, TestValidateRejectsNillableElementKeyFields, TestValidateIdentityPrefixWildcardsAndDescendantFields, TestValidateIdentityConstraintsUseBuiltInValueEquality, TestValidateAnonymousElementTypes, TestValidateSubstitutionGroupMembers, TestSubstitutionGroupMemberInheritsHeadType, TestValidateBlockedSubstitutionGroupDerivation, TestValidateIDAndIDREFSemantics, TestValidateSimpleContentDerivedFromComplexType, TestValidateAnyTypes, TestValidateInlineListAndUnionMembers, TestValidateSimpleContentRestrictionFacets, TestValidateSimpleContentRestrictionOfEmptiableMixedContent, TestValidateFixedValuesForRestrictedSimpleContent | 0 |
| `validate/value_internal_test.go` | test; TestInlineSimpleValueEqualitySelection, TestNamedAndContextualRestrictionValueEquality, TestTypeDerivationMethodsFollowBuiltInSimpleTypeChains, TestDurationValueParsingBoundaries, TestDurationComparisonPreservesArbitraryPrecision, TestCalendarComparisonBoundaries, TestCalendarComparisonUsesXMLSchemaValueSpace, TestContextReaderChecksCancellation, TestValidationPrimitiveDecisionTables, TestValidateElementFixedConstraintBranches, TestIdentityAndFacetPrimitiveDecisionTables, TestValidateIDReferencesBranches, TestParseXMLFloatSpecialValues | 0 |
| `validate/value_tail_internal_test.go` | test; TestSimpleValueEqualityRejectsInvalidOperands, TestIdentityFieldValueBoundaryBranches, TestCanonicalIdentityValueBoundaryBranches, TestSimpleContextAndFacetBoundaryBranches, TestNumericFacetBoundaryBranches | 0 |
| `xsdtest/internal_test.go` | test; TestRunPropagatesAbsolutePathFailure, TestRunReportsMetadataAndRequiredCaseFailures, TestRunRejectsUnreadableAndMalformedMetadata, TestRunHonorsCancellationBetweenGroups, TestCompileDocumentsRejectsMissingAndMalformedInputs, TestSuiteResolverRejectsUnsafeRequests, TestReferenceHelpersRejectUnsafeLexicalForms | 0 |
| `xsdtest/xsts.go` | source | 11 |
| `xsdtest/xsts_test.go` | test; TestOfficialXSTS, TestRunExecutesSchemaAndInstanceExpectations, TestRunClassifiesHintFixtureRefusalsAsErrors, TestRunConfinesSymlinkReadsToSuiteRoot, TestRunRejectsReferencesOutsideSuiteRoot, TestRunDoesNotTreatFixtureReadRefusalAsInvalidSchema, TestRunCompilesAllSchemaDocumentsInAGroup, TestRunCompilesSchemasHintedByAnInstance, TestRunExcludesQueriedExpectations | 0 |

Total: 116 tracked Go files, 29 production files and 87 test files.

## Source-traced contracts

| Owner / consumers | Contract and preserved evidence |
| --- | --- |
| Root model / parser, compiler, serializer, WSDL consumers | Model slices/maps/pointers are mutable; QName identity is namespace plus local part, not prefix. Optional value/occurrence flags distinguish explicit empty/zero from absent syntax. Parse borrows stable bytes with non-nil context and returns an owned model; no external resolution. Namespace copies and string occurrences charge separate per-document allowances before owned amplification. Annotation raw spans are admitted before capture. Existing admission, placement, annotations, QName, occurrence and privacy tests protect these distinctions. |
| Serializer / mutable models and builder | Reflective preflight detects active-path cycles while permitting repeated acyclic references; structure/collection work, nesting and output bytes are independent. Output is deterministic, not canonical XML or original lexical bytes. Supported fragments are checked; schema validity still belongs to compilation. The synchronous public API returns no bytes on error and has no context. Existing cyclic-model, output limit, fragment and round-trip tests remain unchanged. |
| Builder / application-owned construction | Schema owns mutable accumulation and duplicate symbol spaces, with no concurrency safety. AddComplexType recursively clones caller structure before Build's bounded marshal/compiler, so finite acyclic caller input is required; no separate clone context/budget is implemented. AddSimpleRestriction copies only the facet slice and retains nested namespace maps/annotation pointers. Nested complex-type clone tests cover their isolated copies; simple-type cycle preservation is intentional so later serialization can reject the cycle. These distinctions must not become a blanket deep-copy claim. |
| Compiler / validators, WSDL consumers | Per-call graph state/caches are isolated; retained custom resolver must support concurrent calls. Cache parsing by URI differs from effective compilation key (URI, namespace) for chameleon includes. Returned resolver identity is bounded and compared before acceptance; locationless imports receive explicit catalog identity. Graph references, content, components, particles/copy work and per-document parsing budgets are independent. Failure publishes no Set and observed cancellation takes precedence. Set accessors clone mutable components without per-call context/copy budget; compiled projections are not lossless source documents. Composition, clone, cancellation, ambiguity, substitution, derivation and facet tests are preserved. |
| Datatype / compiler and validator | Decimal/integer parsing is exact, capped at 1 MiB before whitespace trimming, with zero-valued immutable numbers and copied private magnitudes before arithmetic. Decimal removes insignificant fractional zeros without rounding. Direct DecimalFacets borrows bound pointers/slices; nonpositive TotalDigits disables checking, nil versus zero FractionDigits differs. Standalone lexical checks do not resolve QName prefixes or universally cover all built-ins. Duration/calendar comparisons use partial orders and arbitrary-precision values; negative-year floor division is necessary. Four fixed duration reference dates, no year zero and absent-timezone intervals retain existing tests/spec scope. Detailed errors can echo lexical data, not fixed redacted categories. |
| Pattern translation / compiler and validator | Fixed XSD block inventory differs from standard toolchain Unicode categories. Translation caps source/translated bytes and subtraction nesting; nil context means no cancellation. Owned work checkpoints precede class admission and protect shared set algebra; sorting/regexp compilation/matching are synchronous. Current whole-value anchoring and supported dialect tests are preserved, not rewritten into new normative policy or broader conformance claims. |
| Resolver / compiler | Memory/Catalog admit count and cumulative input bytes before map allocation and URI work; successful memory bytes are fresh copies. Catalog/Chain retain collaborators and inherit their lifetime/concurrency. Only not-found advances Chain; cancellation after collaborator return prevents fallback. File owns an opt-in os.Root with caller Close, URI/content caps and opened-file cleanup; blocked reads are not interrupted. Capability confinement is not a hardlink/mount/filesystem sandbox. Fixed built-in error categories retain trusted causes and discard partial resources. Exact-boundary, unsafe URI, symlink, cancellation, privacy and classification tests remain intact. |
| Validator / compiled Set | Byte, reader and Node-tree inputs converge on one per-call engine. Readers remain caller-owned and must cooperate with blocked-read cancellation. Tree cloning admits cardinalities and payload before allocation; supplied per-node namespace maps are full scopes, not parent deltas. Counts, strings, namespace copies, identity values and estimated XPath work are separately bounded. Refusal/cancellation returns no partial Result. Diagnostics remain structured trusted data despite supported default redaction. Identity tuples compare normalized value-space representations, not raw spelling. Reader/tree parity, context, numeric facets, QName, wildcard/model groups, identity and private accounting tests remain intact. |
| errsafe / compiler and built-in resolvers | Exact standard context sentinels remain unchanged. Other causes are retained without evaluating delegated format/JSON hooks. Unwrapping, explicit diagnostic fields and invalid fmt verb/type combinations remain trusted-only boundaries; no uniform redaction claim is made for builder/datatype/XSTS detail. |
| xsdtest / trusted XSTS maintainers | Run owns one root capability for metadata and all fixture owners; trusted file sizes/initial XML Unmarshal are not bounded like tenant inputs. Accepted valid/invalid expectations are distinguished from excluded/skipped cases. Individual mismatches populate Report rather than fail Run; cancellation at group boundaries may return an accumulated report. Detailed projections expose metadata/causes. Pinned XSTS/provenance, Java differential and original failed assurance records remain immutable and are not rerun here. |

## Variants, provenance and assurance

All 87 original test files are included above, including helpers for redacted
fmt/JSON/slog output, callback causes, cancellation counters, clones, numeric
boundaries, schema grammar and saved hostile inputs. Four fuzz targets now state
their actual panic-safety or round-trip assertion boundaries; no new oracle or
campaign is introduced. Four embedded-fixture benchmarks describe measured
operations and setup exclusions; no performance claim is added. The two
benchmark go:embed annotations remain unchanged. No build-tag/platform-specific
Go or generated Go producer is present; Windows path behavior is exercised by
existing platform-aware helpers/tests, not a newly qualified runtime claim.
Hand-maintained Unicode block tables do not falsely claim generation.

`api/v2.export` is the immutable apidiff producer baseline, not a GoDoc prose
snapshot. `.golib.yaml` and CI both consume Tools v2.3.0 at the current immutable
pin; only the stale current CONTRIBUTING prose is aligned to that existing fact.
External XSTS/schema fixtures and scripts are not edited or regenerated. Native
provenance checking is offline structural admission only; online resource
refresh and standards/differential qualification are out of scope. All 13
selected normative decisions/history prefixes/source pins remain unchanged.

## Questions and source limitations

No source-backed rationale establishes why builder recursive cloning has no
independent admission/cancellation owner before Build; finite acyclic trusted
input is documented rather than claiming hostile-graph safety or repairing it.
Builder shallow restriction-facet retention is explicit, not invented isolation.
Nil context is unsupported at parser/compiler/validator/resolver entry points
because their work dereferences it, while pattern helpers intentionally accept
nil as context-free use; no new panic or parser probe is executed. Total process
memory, pre-token decoder allocation and noncooperative reader/resolver work are
not bounded/interrupted by package context/owned-work checks. Set projections
are not lossless lexical source models. Standalone datatype helpers lack one
shared input/cancellation budget; callers retain admission responsibility.
XSTS fixture ingestion is trusted test infrastructure, not arbitrary ingestion.
Existing normative scope, original failed security/coverage results and source
qualification questions are preserved rather than reinterpreted.

## Selected verification

Passed under the coordinator's exclusive grant with Go 1.27.2 on darwin/arm64:
all 116 files formatted and preserved exact AST/raw executable tokens/literals
and all six directives/security annotations without semicolon normalization.
All 552 comment groups rendered as HTML and all eight native `go doc -all`
packages rendered. The single actual output-bearing Example passed under its
existing root TestMain. Authenticated Tools v2.3.0 docs/spelling/local links
passed (106 links, zero errors, two existing exclusions); offline specification
validation passed (one module, 13 decisions), as did the existing provenance
script with remote verification disabled. This is native local structure and
selected Example evidence, not a new conformance/security/release verdict.

A final three-file correction retained exact AST/raw tokens/literals and all
annotations, rendered 325 comment groups, and refreshed root/compile GoDocs.
The other six renders and unchanged Example/docs/13 decisions/provenance retain
the original results above rather than being rerun. Current objective source
checker reports zero findings; this is structural admission, not semantic or
standards proof. Both bounded runs removed their exact three caches, helpers,
rendered outputs and logs, including read-only downloaded module directories.

Ordinary complete-diff review reconciled identity-constraint versus facet kinds,
complex-type derivation blocking, global attributes and independent parser
limits against their owners. No executable, authority, dependency, fixture or
API-export change is present. Pinned full changed-file, added-content and staged
secret scans each passed with zero findings; their exact disposable snapshots
and reports were removed. Publication and hosted CI remain coordinator-owned.

## Objective Enforcement

The checksum-pinned public Tools v2.4.0 verifier now supplies
`make source-docs-check` as a prerequisite of the existing local and CI Make
routes. Ordinary shared CI selects strict source-documentation checking and
requires its result. Native prerequisites, operation budgets, security and
release controls are retained. Hosted exact-head evidence is recorded by the
coordination ledger; configuration alone does not establish green delivery.

Choose the next guide in the [documentation index](README.md)
or return to the [repository entry](../README.md) for installation and ownership.
