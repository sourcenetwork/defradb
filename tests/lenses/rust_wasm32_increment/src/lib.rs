// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

use std::sync::RwLock;
use std::error::Error;
use std::fmt;
use serde::Deserialize;
use lens_sdk::StreamOption;
use lens_sdk::error::LensError;

lens_sdk::define!(PARAMETERS: Parameters, try_transform, try_inverse);

#[derive(Clone, PartialEq, Eq, PartialOrd, Ord, Debug, Hash)]
enum ModuleError {
    PropertyNotFoundError{requested: String},
    PropertyNotNumberError{requested: String},
    InvalidIncrementValueError,
}

impl Error for ModuleError { }

impl fmt::Display for ModuleError {
    fn fmt(&self, f: &mut fmt::Formatter) -> fmt::Result {
        match &*self {
            ModuleError::PropertyNotFoundError { requested } =>
                write!(f, "The requested property was not found. Requested: {}", requested),
            ModuleError::PropertyNotNumberError { requested } =>
                write!(f, "The requested property is not a number. Requested: {}", requested),
            ModuleError::InvalidIncrementValueError =>
                write!(f, "The increment value must be a number"),
        }
    }
}

#[derive(Deserialize, Clone)]
pub struct Parameters {
    pub field: String,
    pub value: serde_json::Value,
}

static PARAMETERS: RwLock<Option<Parameters>> = RwLock::new(None);

fn try_transform(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<serde_json::Value>>>,
) -> Result<StreamOption<serde_json::Value>, Box<dyn Error>> {
    apply_increment(iter, |current, increment| current + increment)
}

fn try_inverse(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<serde_json::Value>>>,
) -> Result<StreamOption<serde_json::Value>, Box<dyn Error>> {
    apply_increment(iter, |current, increment| current - increment)
}

fn apply_increment(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<serde_json::Value>>>,
    operation: fn(i64, i64) -> i64,
) -> Result<StreamOption<serde_json::Value>, Box<dyn Error>> {
    let params = PARAMETERS.read()?
        .clone()
        .ok_or(LensError::ParametersNotSetError)?;

    let increment = params.value.as_i64()
        .ok_or(ModuleError::InvalidIncrementValueError)?;

    for item in iter {
        let mut input = match item? {
            Some(v) => v,
            None => return Ok(StreamOption::None),
        };

        let obj = input.as_object_mut()
            .ok_or(ModuleError::PropertyNotFoundError{requested: params.field.clone()})?;

        let field_value = obj.get_mut(&params.field)
            .ok_or(ModuleError::PropertyNotFoundError{requested: params.field.clone()})?;

        let current_value = field_value.as_i64()
            .ok_or(ModuleError::PropertyNotNumberError{requested: params.field.clone()})?;

        let new_value = operation(current_value, increment);

        obj.insert(params.field.clone(), serde_json::Value::Number(new_value.into()));

        return Ok(StreamOption::Some(input))
    }

    Ok(StreamOption::EndOfStream)
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::json;
    use serial_test::serial;

    // Note: All tests use #[serial] because they share the global PARAMETERS static variable.

    #[test]
    #[serial]
    fn test_try_transform_increments_value() {
        let field_name = "count".to_string();
        let mut ptr = PARAMETERS.write().unwrap();
        *ptr = Some(Parameters {
            field: field_name.clone(),
            value: json!(5),
        });
        drop(ptr);

        let input_doc = json!({
            field_name.clone(): 30
        });

        let input = [Ok(Some(input_doc))];
        let mut it = input.into_iter();

        let result = try_transform(&mut it).unwrap();

        let expected_result = json!({
            field_name: 35
        });

        assert_eq!(result, StreamOption::Some(expected_result));
    }

    #[test]
    #[serial]
    fn test_try_transform_with_negative_increment() {
        let field_name = "score".to_string();
        let mut ptr = PARAMETERS.write().unwrap();
        *ptr = Some(Parameters {
            field: field_name.clone(),
            value: json!(-10),
        });
        drop(ptr);

        let input_doc = json!({
            field_name.clone(): 50
        });

        let input = [Ok(Some(input_doc))];
        let mut it = input.into_iter();

        let result = try_transform(&mut it).unwrap();

        let expected_result = json!({
            field_name: 40
        });

        assert_eq!(result, StreamOption::Some(expected_result));
    }

    #[test]
    #[serial]
    fn test_try_transform_handles_empty_iterator() {
        let mut ptr = PARAMETERS.write().unwrap();
        *ptr = Some(Parameters {
            field: "test_field".to_string(),
            value: json!(1),
        });
        drop(ptr);

        let input: [lens_sdk::Result<Option<serde_json::Value>>; 0] = [];
        let mut it = input.into_iter();

        let result = try_transform(&mut it).unwrap();

        assert_eq!(result, StreamOption::EndOfStream);
    }

    #[test]
    #[serial]
    fn test_try_transform_handles_none_input() {
        let mut ptr = PARAMETERS.write().unwrap();
        *ptr = Some(Parameters {
            field: "test_field".to_string(),
            value: json!(1),
        });
        drop(ptr);

        let input = [Ok(None)];
        let mut it = input.into_iter();

        let result = try_transform(&mut it).unwrap();

        assert_eq!(result, StreamOption::None);
    }

    #[test]
    #[serial]
    fn test_try_inverse_decrements_value() {
        let field_name = "balance".to_string();
        let mut ptr = PARAMETERS.write().unwrap();
        *ptr = Some(Parameters {
            field: field_name.clone(),
            value: json!(7),
        });
        drop(ptr);

        let input_doc = json!({
            field_name.clone(): 35
        });

        let input = [Ok(Some(input_doc))];
        let mut it = input.into_iter();

        let result = try_inverse(&mut it).unwrap();

        let expected_result = json!({
            field_name: 28
        });

        assert_eq!(result, StreamOption::Some(expected_result));
    }

    #[test]
    #[serial]
    fn test_transform_then_inverse_roundtrip() {
        let field_name = "value".to_string();
        let original_value = 100;
        let increment_amount = 42;

        // Transform: 100 + 42 = 142
        let mut ptr = PARAMETERS.write().unwrap();
        *ptr = Some(Parameters {
            field: field_name.clone(),
            value: json!(increment_amount),
        });
        drop(ptr);

        let input_doc_1 = json!({
            field_name.clone(): original_value
        });
        let input_1 = [Ok(Some(input_doc_1))];
        let mut it_1 = input_1.into_iter();

        let transform_result = try_transform(&mut it_1).unwrap();

        if let StreamOption::Some(transformed_doc) = transform_result {
            assert_eq!(transformed_doc.get(&field_name).unwrap(), &json!(142));

            // Inverse: 142 - 42 = 100 (reuse same parameters)
            let input_2 = [Ok(Some(transformed_doc))];
            let mut it_2 = input_2.into_iter();

            let inverse_result = try_inverse(&mut it_2).unwrap();

            let expected_result = json!({
                field_name: original_value
            });

            assert_eq!(inverse_result, StreamOption::Some(expected_result));
        } else {
            panic!("Transform should return Some");
        }
    }
}
