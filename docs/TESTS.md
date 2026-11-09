# Tests

| Test | Rule |
|---|---|
| [TestAssuranceZeroValueIsNone](../tests/assurance_test.go) | AssuranceZeroValueIsNone |
| [TestAssuranceIsOrdered](../tests/assurance_test.go) | AssuranceIsOrdered |
| [TestAssuranceString](../tests/assurance_test.go) | AssuranceString |
| [TestCodecRoundTrip](../tests/codec_test.go) | CodecRoundTrip |
| [TestCodecAllocations](../tests/codec_test.go) | CodecAllocations |
| [TestDefinition](../tests/definition_test.go) | Definition |
| [TestDefinition_AllowZeroIsClosed](../tests/definition_test.go) | Definition AllowZeroIsClosed |
| [TestDefinition_AllowHas](../tests/definition_test.go) | Definition AllowHas |
| [TestKey_JoinSplitRoundTrip](../tests/definition_test.go) | Key JoinSplitRoundTrip |
| [TestKey_SplitWrongCountNotOK](../tests/definition_test.go) | Key SplitWrongCountNotOK |
| [TestKey_SplitEmptyPartNotOK](../tests/definition_test.go) | Key SplitEmptyPartNotOK |
| [TestKindOf_Declared](../tests/error_test.go) | KindOf Declared |
| [TestKindOf_PlainIsInternal](../tests/error_test.go) | KindOf PlainIsInternal |
| [TestFieldOf_RequiredField](../tests/error_test.go) | FieldOf RequiredField |
| [TestFieldOf_TooLong](../tests/error_test.go) | FieldOf TooLong |
| [TestFieldOf_NotAField](../tests/error_test.go) | FieldOf NotAField |
| [TestErrorKind_String](../tests/error_test.go) | ErrorKind String |
| [TestFieldOf_AnyErrorThatNamesItsField](../tests/error_test.go) | FieldOf AnyErrorThatNamesItsField |
| [TestFieldTypeString](../tests/field_test.go) | FieldTypeString |
| [TestFieldDisplayTexts](../tests/field_test.go) | FieldDisplayTexts |
| [TestBaseKinds](../tests/field_test.go) | BaseKinds |
| [TestFieldValidate_FailClosed](../tests/field_test.go) | FieldValidate FailClosed |
| [TestFieldValidate_ExplicitWhitelistReplacesTextFloor](../tests/field_test.go) | FieldValidate ExplicitWhitelistReplacesTextFloor |
| [TestFieldValidate_FloorKeptWithoutPositiveWhitelist](../tests/field_test.go) | FieldValidate FloorKeptWithoutPositiveWhitelist |
| [TestFieldValidate_NotNull_EmptyValue](../tests/field_test.go) | FieldValidate NotNull EmptyValue |
| [TestFieldZeroValue](../tests/field_test.go) | FieldZeroValue |
| [TestFieldOmitEmpty](../tests/field_test.go) | FieldOmitEmpty |
| [TestFieldConstraints](../tests/field_test.go) | FieldConstraints |
| [TestFieldValidate](../tests/field_test.go) | FieldValidate |
| [TestActionConstantValues](../tests/field_test.go) | ActionConstantValues |
| [TestValidateFieldsRecursive](../tests/field_test.go) | ValidateFieldsRecursive |
| [TestReadValuesAllTypes](../tests/field_test.go) | ReadValuesAllTypes |
| [TestIsZeroPtrAllTypes](../tests/field_test.go) | IsZeroPtrAllTypes |
| [TestReadStringPtr](../tests/field_test.go) | ReadStringPtr |
| [TestValidateFieldsWithOnlyFielder](../tests/field_test.go) | ValidateFieldsWithOnlyFielder |
| [TestValidateFieldsActions](../tests/field_test.go) | ValidateFieldsActions |
| [TestFielderSliceIsTraversalOnly](../tests/field_test.go) | FielderSliceIsTraversalOnly |
| [TestRefKind](../tests/field_test.go) | RefKind |
| [TestNilRefBackstop](../tests/field_test.go) | NilRefBackstop |
| [TestSliceNeedsNoSchema](../tests/fielder_slice_test.go) | SliceNeedsNoSchema |
| [TestJoinItem](../tests/item_test.go) | JoinItem |
| [TestValidateItem_Valid](../tests/item_test.go) | ValidateItem Valid |
| [TestValidateItem_NoMarksIsValid](../tests/item_test.go) | ValidateItem NoMarksIsValid |
| [TestValidateItem_TwoLabels](../tests/item_test.go) | ValidateItem TwoLabels |
| [TestValidateItem_TwoBadges](../tests/item_test.go) | ValidateItem TwoBadges |
| [TestValidateItem_DescriptionWithoutLabel](../tests/item_test.go) | ValidateItem DescriptionWithoutLabel |
| [TestValidateItem_NoPK](../tests/item_test.go) | ValidateItem NoPK |
| [TestValidateItem_RefAndOptions](../tests/item_test.go) | ValidateItem RefAndOptions |
| [TestFieldPermitted_ExtendsTextKindCharset](../tests/kind_permitted_override_test.go) | FieldPermitted ExtendsTextKindCharset |
| [TestFieldPermitted_StillRejectsUnlisted](../tests/kind_permitted_override_test.go) | FieldPermitted StillRejectsUnlisted |
| [TestPermittedValidate](../tests/permitted_test.go) | PermittedValidate |
| [TestPermittedErrorMessages](../tests/permitted_test.go) | PermittedErrorMessages |
| [TestPermittedRecursiveStartWith](../tests/permitted_test.go) | PermittedRecursiveStartWith |
| [TestPermitted_NoHTML_BlocksInjection](../tests/permitted_test.go) | Permitted NoHTML BlocksInjection |
| [TestPermitted_NoHTML_AllowsNormal](../tests/permitted_test.go) | Permitted NoHTML AllowsNormal |
| [TestWidget_StandardPrinciples_RejectsHTML](../tests/permitted_test.go) | Widget StandardPrinciples RejectsHTML |
| [TestAccessZeroValueIsGuarded](../tests/rbac_test.go) | AccessZeroValueIsGuarded |
| [TestClosedByDefault](../tests/rbac_test.go) | ClosedByDefault |
| [TestActionsAreASet](../tests/rbac_test.go) | ActionsAreASet |
| [TestGrantMatches](../tests/rbac_test.go) | GrantMatches |
| [TestFullAccessIsNeverGrantedImplicitly](../tests/rbac_test.go) | FullAccessIsNeverGrantedImplicitly |
| [TestActionRoundTripsAsLetters](../tests/rbac_test.go) | ActionRoundTripsAsLetters |
| [TestParseActionIgnoresOrder](../tests/rbac_test.go) | ParseActionIgnoresOrder |
| [TestParseActionRejectsUnknownVerb](../tests/rbac_test.go) | ParseActionRejectsUnknownVerb |
| [TestAllowedDelegates](../tests/rbac_test.go) | AllowedDelegates |
| [TestRolesFor](../tests/rbac_test.go) | RolesFor |
| [TestAccessRendersAsAWordNotANumber](../tests/rbac_test.go) | AccessRendersAsAWordNotANumber |
| [TestResourceOf_Table](../tests/resource_test.go) | ResourceOf Table |
| [TestResourceOf_NotATable](../tests/resource_test.go) | ResourceOf NotATable |
| [TestResourceOf_Nil](../tests/resource_test.go) | ResourceOf Nil |
| [TestFeatureResource](../tests/resource_test.go) | FeatureResource |
| [TestValidateInput_SkipsServerOnly](../tests/server_only_test.go) | ValidateInput SkipsServerOnly |
| [TestValidateInput_StillChecksClientFields](../tests/server_only_test.go) | ValidateInput StillChecksClientFields |
| [TestValidateFields_ChecksServerOnly](../tests/server_only_test.go) | ValidateFields ChecksServerOnly |
| [TestInstantKinds_StoreAsIntAndValidateLikeInt](../tests/time_kind_test.go) | InstantKinds StoreAsIntAndValidateLikeInt |
| [TestVector_StorageIsBlob](../tests/vector_test.go) | Vector StorageIsBlob |
| [TestVector_Dim](../tests/vector_test.go) | Vector Dim |
| [TestVector_ValidateOK](../tests/vector_test.go) | Vector ValidateOK |
| [TestVector_ValidateWrongDim](../tests/vector_test.go) | Vector ValidateWrongDim |
| [TestVector_ValidateNotMultipleOfFour](../tests/vector_test.go) | Vector ValidateNotMultipleOfFour |
| [TestVector_ValidateEmpty](../tests/vector_test.go) | Vector ValidateEmpty |
| [TestVector_ZeroPtr](../tests/vector_test.go) | Vector ZeroPtr |
| [TestBlob_StillNotDimensional](../tests/vector_test.go) | Blob StillNotDimensional |
| [TestVerb_OfEachAction](../tests/verb_test.go) | Verb OfEachAction |
| [TestVerb_OfASetOrNothingIsEmpty](../tests/verb_test.go) | Verb OfASetOrNothingIsEmpty |
| [TestParseVerb_RoundTrips](../tests/verb_test.go) | ParseVerb RoundTrips |
| [TestParseVerb_UnknownIsNotAnAction](../tests/verb_test.go) | ParseVerb UnknownIsNotAnAction |
| [TestField_IsRevision](../tests/field_revision_test.go) | Field IsRevision |
