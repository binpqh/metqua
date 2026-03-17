import pandas as pd
import numpy as np
from sklearn.compose import ColumnTransformer
from sklearn.preprocessing import OneHotEncoder, LabelEncoder

dataset = pd.read_csv("machine-learning/0-preprocessing/titanic.csv")

# X keeps PassengerId
X = dataset.drop("Survived", axis=1)
y = dataset["Survived"]

ohe = OneHotEncoder(
    categories=[
        [1, 2, 3],
        ["female", "male"],
        ["C", "Q", "S"]
    ],
    handle_unknown="ignore"
)

ct = ColumnTransformer(
    transformers=[
        ("encoder", ohe, ["Pclass", "Sex", "Embarked"])
    ],
    remainder="passthrough"
)

X = np.array(ct.fit_transform(X))

le = LabelEncoder()
y = le.fit_transform(y)

print(X)
print(y)
print(X.shape)