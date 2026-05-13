import { Image, StyleSheet, View } from 'react-native';
import { theme } from '../constants/theme';

type Props = {
  uri: string;
  size?: number;
};

export const ProfileAvatar = ({ uri, size = 104 }: Props) => {
  return (
    <View style={[styles.wrap, { width: size, height: size, borderRadius: size / 2 }]}>
      <Image source={{ uri }} style={[styles.image, { borderRadius: size / 2 }]} />
    </View>
  );
};

const styles = StyleSheet.create({
  wrap: {
    padding: 3,
    borderWidth: 2,
    borderColor: theme.colors.primary,
    backgroundColor: theme.colors.surface,
  },
  image: {
    width: '100%',
    height: '100%',
  },
});
