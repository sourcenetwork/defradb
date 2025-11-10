use std::collections::HashMap;
use std::sync::RwLock;
use std::error::Error;
use serde::Deserialize;
use lens_sdk::StreamOption;
use lens_sdk::error::LensError;

lens_sdk::define!(PARAMETERS: Parameters, try_transform);

#[derive(Deserialize, Clone)]
pub struct Parameters {
    pub src: String,
    pub dst: String,
}

static PARAMETERS: RwLock<Option<Parameters>> = RwLock::new(None);

fn try_transform(
    iter: &mut dyn Iterator<Item = lens_sdk::Result<Option<HashMap<String, serde_json::Value>>>>,
) -> Result<StreamOption<HashMap<String, serde_json::Value>>, Box<dyn Error>> {
    let params = PARAMETERS.read()?
        .clone()
        .ok_or(LensError::ParametersNotSetError)?;

    let mut values = Vec::<f64>::new();
    for item in iter {
        let mut input = match item? {
            Some(v) => v,
            None => break,
        };
        
        let value = input.get_mut(&params.src)
            .unwrap()
            .clone();

        values.push(value.as_u64().unwrap() as f64);
    }

    if values.len() == 0 {
        return Ok(StreamOption::EndOfStream);
    }

    let mut sum = 0f64;
    let count = values.len() as f64;
    for v in &values {
        sum = sum + v;
    }

    let mean = sum / count;

    let mut sum_dev = 0f64;
    for v in &values {
        let dev = v - mean;
        let dev_sqd = dev*dev;
        sum_dev = sum_dev + dev_sqd;
    }

    let variance = sum_dev / count;
    let std_dev = variance.sqrt();

    let mut result = HashMap::<String, serde_json::Value>::new();
    result.insert(params.dst, std_dev.into());

    Ok(StreamOption::Some(result))
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_try_transform_pass() {
        let src_f = "SRC".to_string();
        let mut ptr = PARAMETERS.write().unwrap();
        *ptr = Some(Parameters{
            src: src_f.clone(),
            dst: "DST".to_string(),
        });
        drop(ptr);

        let mut input_doc_1 = HashMap::<String, serde_json::Value>::new();
        input_doc_1.insert(src_f.clone(), 10.into());
        let mut input_doc_2 = HashMap::<String, serde_json::Value>::new();
        input_doc_2.insert(src_f.clone(), 14.into());

        let input = [
            Ok(
                Some(
                    input_doc_1,
                ),
            ),
            Ok(
                Some(
                    input_doc_2,
                ),
            ),
        ];

        let mut it = input.into_iter();

        let result = try_transform(&mut it).unwrap();

        let mut expected_result = HashMap::<String, serde_json::Value>::new();
        expected_result.insert("DST".to_string(), 2f64.into());

        assert_eq!(
            result,
            StreamOption::Some(expected_result.clone()),
        );

        let result_2 = try_transform(&mut it).unwrap();
        assert_eq!(
            result_2,
            StreamOption::EndOfStream,
        );
    }
}
